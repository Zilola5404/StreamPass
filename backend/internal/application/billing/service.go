package billing

import (
	"context"
	"log/slog"
	"time"

	"streampass/backend/internal/domain/subscription"
	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/idgen"
	"streampass/shared/logger"
)

// Plan describes a sellable subscription plan (month / year). Loaded from
// config (spec: "Запрещено использовать hardcode"), not a constant, so
// price/period can change without a code deploy.
type Plan struct {
	Code       string // "month" | "quarter" | "year"
	Title      string
	AmountRUB  int64 // RUB for card providers; Stars for Telegram mode
	PeriodDays int
	Currency   string // RUB | XTR
}

// TelegramInvoicer creates Stars invoice links keyed by our payment id.
type TelegramInvoicer interface {
	CreateInvoiceForPayment(ctx context.Context, paymentID, planCode string) (invoiceLink string, stars int, periodDays int, title string, err error)
	MarkSucceededCharge(providerPaymentID string)
}

// Service implements the Billing use cases.
type Service struct {
	users     user.Repository
	payments  subscription.PaymentRepository
	provider  PaymentProvider
	telegram  TelegramInvoicer
	plans     []Plan
	clock     Clock
	log       *logger.Logger
}

// NewService wires the Billing service via constructor injection.
func NewService(users user.Repository, payments subscription.PaymentRepository, provider PaymentProvider, plans []Plan, clock Clock, log *logger.Logger) *Service {
	if len(plans) == 0 {
		plans = []Plan{{Code: "month", Title: "Месяц", AmountRUB: 299, PeriodDays: 30, Currency: "RUB"}}
	}
	for i := range plans {
		if plans[i].Currency == "" {
			plans[i].Currency = "RUB"
		}
	}
	return &Service{users: users, payments: payments, provider: provider, plans: plans, clock: clock, log: log.With("billing_service")}
}

// SetTelegramInvoicer enables Stars invoice creation for POST /payments.
func (s *Service) SetTelegramInvoicer(inv TelegramInvoicer) { s.telegram = inv }

// ListPlans returns available tariffs for GET /plans.
func (s *Service) ListPlans() []Plan {
	return append([]Plan(nil), s.plans...)
}

func (s *Service) resolvePlan(code string) (Plan, error) {
	if code == "" {
		return s.plans[0], nil
	}
	for _, p := range s.plans {
		if p.Code == code {
			return p, nil
		}
	}
	return Plan{}, apperrors.New(apperrors.CodeInvalidInput, "unknown plan").
		WithDetails(map[string]any{"field": "plan_code", "plan_code": code})
}

// CreatePayment starts a new payment for the user's subscription.
func (s *Service) CreatePayment(ctx context.Context, userID user.ID, planCode string) (confirmationURL string, err error) {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return "", err
	}
	plan, err := s.resolvePlan(planCode)
	if err != nil {
		return "", err
	}

	paymentID := idgen.New()
	provider := "yookassa"
	currency := plan.Currency
	if currency == "" {
		currency = "RUB"
	}
	amount := plan.AmountRUB
	periodDays := plan.PeriodDays
	providerPaymentID := paymentID
	var url string

	if s.telegram != nil {
		// Persist first so pre_checkout FindByID(payload) succeeds.
		payment := &subscription.Payment{
			ID:         paymentID,
			UserID:     string(userID),
			ProviderID: paymentID,
			AmountRUB:  amount,
			PeriodDays: periodDays,
			Status:     subscription.PaymentPending,
			CreatedAt:  s.clock.Now(),
			Provider:   "telegram",
			Currency:   "XTR",
			Tariff:     plan.Code,
		}
		if err := s.payments.Create(ctx, payment); err != nil {
			s.log.Error(ctx, err)
			return "", apperrors.Wrap(apperrors.CodeInternal, "failed to record pending payment", err)
		}
		link, stars, days, title, err := s.telegram.CreateInvoiceForPayment(ctx, paymentID, plan.Code)
		if err != nil {
			s.log.Error(ctx, err)
			return "", apperrors.Wrap(apperrors.CodePaymentFailed, "failed to create telegram invoice", err)
		}
		url = link
		_ = title
		_ = days
		s.log.Info(ctx, "telegram invoice created",
			slog.String("payment_id", paymentID),
			slog.String("plan", plan.Code),
			slog.Int("stars", stars),
		)
		return url, nil
	}

	id, confURL, err := s.provider.CreatePayment(ctx, string(userID), plan.AmountRUB, "StreamPass "+plan.Title)
	if err != nil {
		s.log.Error(ctx, err)
		return "", apperrors.Wrap(apperrors.CodePaymentFailed, "failed to create payment", err)
	}
	providerPaymentID = id
	url = confURL
	provider = s.providerName()

	payment := &subscription.Payment{
		ID:         paymentID,
		UserID:     string(userID),
		ProviderID: providerPaymentID,
		AmountRUB:  amount,
		PeriodDays: periodDays,
		Status:     subscription.PaymentPending,
		CreatedAt:  s.clock.Now(),
		Provider:   provider,
		Currency:   currency,
		Tariff:     plan.Code,
	}
	if err := s.payments.Create(ctx, payment); err != nil {
		s.log.Error(ctx, err)
		return "", apperrors.Wrap(apperrors.CodeInternal, "failed to record pending payment", err)
	}

	return url, nil
}

func (s *Service) providerName() string {
	type named interface{ Name() string }
	if n, ok := s.provider.(named); ok {
		return n.Name()
	}
	return "yookassa"
}

// HandleProviderConfirmed credits a payment after a trusted provider webhook
// (Platega CONFIRMED). Still re-fetches status when FetchPaymentStatus works.
func (s *Service) HandleProviderConfirmed(ctx context.Context, providerPaymentID string) error {
	return s.HandleWebhook(ctx, providerPaymentID)
}

// HandleWebhook processes a payment-provider notification.
func (s *Service) HandleWebhook(ctx context.Context, providerPaymentID string) error {
	payment, err := s.payments.FindByProviderID(ctx, providerPaymentID)
	if err != nil {
		return err
	}
	if payment.Status == subscription.PaymentSucceeded {
		return nil
	}

	status, err := s.provider.FetchPaymentStatus(ctx, providerPaymentID)
	if err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodePaymentFailed, "failed to confirm payment status with provider", err)
	}
	if status != PaymentStatusSucceeded {
		return nil
	}

	return s.creditPayment(ctx, payment)
}

// ActivateByPaymentID credits a pending payment found by our payment id
// (Telegram invoice payload). Idempotent.
func (s *Service) ActivateByPaymentID(ctx context.Context, paymentID, chargeID string, tgUserID *int64) (activeUntil time.Time, err error) {
	payment, err := s.payments.FindByID(ctx, paymentID)
	if err != nil {
		return time.Time{}, err
	}
	if payment.Status == subscription.PaymentSucceeded {
		u, err := s.users.FindByID(ctx, user.ID(payment.UserID))
		if err != nil {
			return time.Time{}, err
		}
		if u.SubscriptionActiveUntil != nil {
			return *u.SubscriptionActiveUntil, nil
		}
		return s.clock.Now(), nil
	}
	if err := s.payments.MarkSucceededByID(ctx, payment.ID, chargeID, tgUserID); err != nil {
		return time.Time{}, err
	}
	if s.telegram != nil {
		s.telegram.MarkSucceededCharge(payment.ProviderID)
	}
	until, err := s.extendFromPayment(ctx, payment)
	if err != nil {
		return time.Time{}, err
	}
	s.log.Info(ctx, "subscription activated",
		slog.String("user_id", payment.UserID),
		slog.String("payment_id", payment.ID),
		slog.Time("until", until),
	)
	return until, nil
}

func (s *Service) creditPayment(ctx context.Context, payment *subscription.Payment) error {
	if err := s.payments.MarkSucceeded(ctx, payment.ProviderID); err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to mark payment succeeded", err)
	}
	_, err := s.extendFromPayment(ctx, payment)
	return err
}

func (s *Service) extendFromPayment(ctx context.Context, payment *subscription.Payment) (time.Time, error) {
	u, err := s.users.FindByID(ctx, user.ID(payment.UserID))
	if err != nil {
		s.log.Error(ctx, err)
		return time.Time{}, apperrors.Wrap(apperrors.CodeInternal, "failed to load user for subscription", err)
	}
	base := s.clock.Now()
	if u.SubscriptionActiveUntil != nil && u.SubscriptionActiveUntil.After(base) {
		base = *u.SubscriptionActiveUntil
	}
	newExpiry := base.Add(time.Duration(payment.PeriodDays) * 24 * time.Hour)
	if err := s.users.ExtendSubscription(ctx, user.ID(payment.UserID), newExpiry); err != nil {
		s.log.Error(ctx, err)
		return time.Time{}, apperrors.Wrap(apperrors.CodeInternal, "failed to extend subscription", err)
	}
	return newExpiry, nil
}

// GetSubscription implements "GET /subscription".
func (s *Service) GetSubscription(ctx context.Context, userID user.ID) (subscription.Info, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return subscription.Info{}, err
	}
	return subscription.NewInfoWithTrial(
		u.SubscriptionActiveUntil,
		u.TrialEndsAt,
		u.EntitlementSource,
		s.clock.Now(),
	), nil
}

// ListPayments implements "GET /payments" history (E06).
func (s *Service) ListPayments(ctx context.Context, userID user.ID) ([]*subscription.Payment, error) {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.payments.ListByUserID(ctx, string(userID))
}

// CancelSubscription implements "отмена" (FS E06).
func (s *Service) CancelSubscription(ctx context.Context, userID user.ID) error {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return err
	}
	s.log.Info(ctx, "subscription cancel acknowledged (access until active_until)")
	return nil
}

// PaymentsRepo exposes the payment repository for Telegram/USDT modules.
func (s *Service) PaymentsRepo() subscription.PaymentRepository { return s.payments }

// UsersRepo exposes users for Telegram/USDT modules.
func (s *Service) UsersRepo() user.Repository { return s.users }

// ClockNow returns current time.
func (s *Service) ClockNow() time.Time { return s.clock.Now() }
