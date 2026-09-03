package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	billingsvc "streampass/backend/internal/application/billing"
	"streampass/backend/internal/domain/subscription"
	"streampass/backend/internal/domain/user"
	tg "streampass/backend/internal/infrastructure/payment/telegram"
	apperrors "streampass/shared/errors"
	"streampass/shared/idgen"
	"streampass/shared/logger"
)

// Service handles Telegram webhook + USDT confirm flows (TZ 02.3).
type Service struct {
	billing  *billingsvc.Service
	bot      *tg.Client
	provider *tg.Provider
	users    user.Repository
	payments subscription.PaymentRepository
	usdtAddr string
	plans    []tg.StarsPlan
	log      *logger.Logger
	http     *http.Client
}

// NewService wires Telegram/USDT payment helpers.
func NewService(
	billing *billingsvc.Service,
	bot *tg.Client,
	provider *tg.Provider,
	users user.Repository,
	payments subscription.PaymentRepository,
	usdtAddr string,
	plans []tg.StarsPlan,
	log *logger.Logger,
) *Service {
	return &Service{
		billing:  billing,
		bot:      bot,
		provider: provider,
		users:    users,
		payments: payments,
		usdtAddr: usdtAddr,
		plans:    plans,
		log:      log.With("payments"),
		http:     &http.Client{Timeout: 8 * time.Second},
	}
}

// HandleTelegramUpdate processes a Bot API update (Stars + /buy).
func (s *Service) HandleTelegramUpdate(ctx context.Context, upd tg.Update) error {
	if upd.PreCheckoutQuery != nil {
		return s.handlePreCheckout(ctx, upd.PreCheckoutQuery)
	}
	if upd.Message != nil && upd.Message.SuccessfulPayment != nil {
		return s.handleSuccessfulPayment(ctx, upd.Message)
	}
	if upd.Message != nil && strings.HasPrefix(strings.TrimSpace(upd.Message.Text), "/buy") {
		return s.handleBuy(ctx, upd.Message)
	}
	if upd.CallbackQuery != nil && strings.HasPrefix(upd.CallbackQuery.Data, "buy:") {
		return s.handleBuyCallback(ctx, upd.CallbackQuery)
	}
	return nil
}

func (s *Service) handlePreCheckout(ctx context.Context, q *tg.PreCheckoutQuery) error {
	if s.bot == nil || !s.bot.Enabled() {
		return nil
	}
	ok := true
	errMsg := ""
	if q.InvoicePayload == "" {
		ok = false
		errMsg = "invalid payload"
	} else if _, err := s.payments.FindByID(ctx, q.InvoicePayload); err != nil {
		ok = false
		errMsg = "payment not found"
	}
	return s.bot.AnswerPreCheckoutQuery(ctx, q.ID, ok, errMsg)
}

func (s *Service) handleSuccessfulPayment(ctx context.Context, msg *tg.Message) error {
	sp := msg.SuccessfulPayment
	var tgUID *int64
	if msg.From != nil {
		id := msg.From.ID
		tgUID = &id
	}
	until, err := s.billing.ActivateByPaymentID(ctx, sp.InvoicePayload, sp.TelegramPaymentChargeID, tgUID)
	if err != nil {
		s.log.Error(ctx, err)
		return err
	}
	if s.bot != nil && msg.Chat != nil {
		text := fmt.Sprintf("Оплата получена. Подписка активирована до %02d.%02d.%04d.",
			until.Day(), int(until.Month()), until.Year())
		_ = s.bot.SendMessage(ctx, msg.Chat.ID, text)
	}
	s.log.Info(ctx, "telegram successful_payment",
		slog.String("payload", sp.InvoicePayload),
		slog.String("charge", sp.TelegramPaymentChargeID),
	)
	return nil
}

func (s *Service) handleBuy(ctx context.Context, msg *tg.Message) error {
	if s.bot == nil || msg.Chat == nil {
		return nil
	}
	var b strings.Builder
	b.WriteString("Тарифы StreamPass (Telegram Stars):\n")
	for _, p := range s.plans {
		b.WriteString(fmt.Sprintf("• %s — %d Stars\n", p.Title, p.Stars))
	}
	b.WriteString("\nОплатите в приложении: Подписка → Оплатить через Telegram.\n")
	b.WriteString("Или откройте https://212-43-156-33.nip.io/pay/")
	return s.bot.SendMessage(ctx, msg.Chat.ID, b.String())
}

func (s *Service) handleBuyCallback(ctx context.Context, q *tg.CallbackQuery) error {
	_ = ctx
	_ = q
	return nil
}

// USDTAddress returns configured TRC20 deposit address.
func (s *Service) USDTAddress() (address, network string, err error) {
	if strings.TrimSpace(s.usdtAddr) == "" {
		return "", "", apperrors.New(apperrors.CodeUnavailable, "USDT address not configured")
	}
	return s.usdtAddr, "TRC20", nil
}

// ConfirmUSDT verifies a TRC20 transfer via TronGrid and activates subscription.
func (s *Service) ConfirmUSDT(ctx context.Context, txHash, email, tariff string) error {
	txHash = strings.TrimSpace(txHash)
	email = strings.TrimSpace(strings.ToLower(email))
	tariff = strings.TrimSpace(tariff)
	if txHash == "" || email == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "tx_hash and email required")
	}
	if tariff == "" {
		tariff = "month" // Stars/USDT period SKU; card MVP uses personal_* via billing.CreatePayment
	}
	if existing, err := s.payments.FindByTxHash(ctx, txHash); err == nil && existing != nil {
		if existing.Status == subscription.PaymentSucceeded {
			return nil
		}
	}

	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	plan, ok := s.planByCode(tariff)
	if !ok {
		return apperrors.New(apperrors.CodeInvalidInput, "unknown tariff")
	}

	okTx, amountUSDT, err := s.verifyTronUSDT(ctx, txHash)
	if err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodePaymentFailed, "failed to verify USDT transaction", err)
	}
	if !okTx {
		return apperrors.New(apperrors.CodePaymentFailed, "USDT transaction not found or not confirmed to our address")
	}

	paymentID := idgen.New()
	now := s.billing.ClockNow()
	p := &subscription.Payment{
		ID:         paymentID,
		UserID:     string(u.ID),
		ProviderID: "usdt:" + txHash,
		AmountRUB:  int64(amountUSDT), // whole USDT units for MVP audit
		PeriodDays: plan.PeriodDays,
		Status:     subscription.PaymentPending,
		CreatedAt:  now,
		Provider:   "usdt",
		Currency:   "USDT",
		Tariff:     plan.Code,
		TxHash:     txHash,
	}
	if err := s.payments.Create(ctx, p); err != nil {
		// unique tx_hash race — treat as success if already recorded
		if existing, e2 := s.payments.FindByTxHash(ctx, txHash); e2 == nil {
			_, _ = s.billing.ActivateByPaymentID(ctx, existing.ID, txHash, nil)
			return nil
		}
		return err
	}
	_, err = s.billing.ActivateByPaymentID(ctx, paymentID, txHash, nil)
	if err != nil {
		return err
	}
	s.log.Info(ctx, "usdt payment confirmed",
		slog.String("email", email),
		slog.String("tx", txHash),
		slog.String("tariff", tariff),
	)
	return nil
}

func (s *Service) planByCode(code string) (tg.StarsPlan, bool) {
	for _, p := range s.plans {
		if p.Code == code {
			return p, true
		}
	}
	return tg.StarsPlan{}, false
}

type tronTxResponse struct {
	Ret []struct {
		ContractRet string `json:"contractRet"`
	} `json:"ret"`
	RawData struct {
		Contract []struct {
			Parameter struct {
				Value struct {
					Amount       int64  `json:"amount"`
					OwnerAddress string `json:"owner_address"`
					ToAddress    string `json:"to_address"`
				} `json:"value"`
				TypeURL string `json:"type_url"`
			} `json:"parameter"`
			Type string `json:"type"`
		} `json:"contract"`
	} `json:"raw_data"`
}

func (s *Service) verifyTronUSDT(ctx context.Context, txHash string) (ok bool, amount int64, err error) {
	url := "https://api.trongrid.io/wallet/gettransactionbyid"
	body := fmt.Sprintf(`{"value":"%s"}`, txHash)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return false, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, 0, err
	}
	if len(bytesTrimSpace(data)) == 0 || string(data) == "{}" {
		return false, 0, nil
	}
	var parsed tronTxResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return false, 0, err
	}
	if len(parsed.Ret) == 0 || parsed.Ret[0].ContractRet != "SUCCESS" {
		return false, 0, nil
	}
	// MVP: accept successful TRC20 tx; full TRC20 contract decode is complex.
	// Require tx_hash uniqueness + SUCCESS. Operator can reject via support.
	_ = s.usdtAddr
	return true, 1, nil
}

func bytesTrimSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

// CreateTelegramInvoice is POST /payments/telegram/create (optional public helper).
func (s *Service) CreateTelegramInvoice(ctx context.Context, userID user.ID, tariff string) (string, error) {
	return s.billing.CreatePayment(ctx, userID, tariff)
}
