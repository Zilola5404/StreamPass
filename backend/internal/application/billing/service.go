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

// Plan describes a sellable subscription plan. Loaded from config
// (BILLING-001: prices never hard-coded in clients).
type Plan struct {
	Code        string // personal_basic | personal_pro | business (+ aliases)
	Title       string
	AmountRUB   int64 // RUB for card providers; Stars for Telegram mode
	PeriodDays  int
	Currency    string // RUB | XTR
	MaxDevices  int
	MaxUsers    int
	Description string
}

// TelegramInvoicer creates Stars invoice links keyed by our payment id.
type TelegramInvoicer interface {
	CreateInvoiceForPayment(ctx context.Context, paymentID, planCode string) (invoiceLink string, stars int, periodDays int, title string, err error)
	MarkSucceededCharge(providerPaymentID string)
}

// Service implements the Billing use cases.
type Service struct {
	users    user.Repository
	payments subscription.PaymentRepository
	orders   subscription.OrderRepository
	provider PaymentProvider
	telegram TelegramInvoicer
	plans    []Plan
	clock    Clock
	log      *logger.Logger
}

// NewService wires the Billing service via constructor injection.
func NewService(
	users user.Repository,
	payments subscription.PaymentRepository,
	orders subscription.OrderRepository,
	provider PaymentProvider,
	plans []Plan,
	clock Clock,
	log *logger.Logger,
) *Service {
	if len(plans) == 0 {
		plans = []Plan{{
			Code: "personal_basic", Title: "Personal Basic",
			AmountRUB: 299, PeriodDays: 30, Currency: "RUB", MaxDevices: 2, MaxUsers: 1,
		}}
	}
	for i := range plans {
		if plans[i].Currency == "" {
			plans[i].Currency = "RUB"
		}
	}
	return &Service{
		users: users, payments: payments, orders: orders,
		provider: provider, plans: plans, clock: clock, log: log.With("billing_service"),
	}
}

// SetTelegramInvoicer enables Stars invoice creation for POST /payments.
func (s *Service) SetTelegramInvoicer(inv TelegramInvoicer) { s.telegram = inv }

// ListPlans returns available tariffs for GET /plans.
// Prefer BILLING-001 canonical codes when present; otherwise return full catalog
// (Telegram Stars month/quarter/year).
func (s *Service) ListPlans() []Plan {
	preferred := []string{"personal_basic", "personal_pro", "business"}
	out := make([]Plan, 0, len(preferred))
	for _, code := range preferred {
		for _, p := range s.plans {
			if p.Code == code {
				out = append(out, p)
				break
			}
		}
	}
	if len(out) > 0 {
		return out
	}
	seen := map[string]bool{}
	out = make([]Plan, 0, len(s.plans))
	for _, p := range s.plans {
		if seen[p.Code] {
			continue
		}
		seen[p.Code] = true
		out = append(out, p)
	}
	return out
}

func (s *Service) resolvePlan(code string) (Plan, error) {
	if code == "" {
		return s.plans[0], nil
	}
	// Normalize legacy aliases → BILLING-001 codes.
	switch code {
	case "basic", "month":
		code = "personal_basic"
	case "pro":
		code = "personal_pro"
	}
	for _, p := range s.plans {
		if p.Code == code {
			return p, nil
		}
	}
	// Fall back to matching aliases still listed in catalog.
	for _, p := range s.plans {
		if p.Code == code || (code == "personal_basic" && (p.Code == "basic" || p.Code == "month")) ||
			(code == "personal_pro" && p.Code == "pro") {
			return p, nil
		}
	}
	return Plan{}, apperrors.New(apperrors.CodeInvalidInput, "unknown plan").
		WithDetails(map[string]any{"field": "plan_code", "plan_code": code})
}

// CreateOrder creates a PENDING order for a plan (BILLING-001 step).
func (s *Service) CreateOrder(ctx context.Context, userID user.ID, planCode string) (*subscription.Order, error) {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return nil, err
	}
	plan, err := s.resolvePlan(planCode)
	if err != nil {
		return nil, err
	}
	order := &subscription.Order{
		ID:         idgen.New(),
		UserID:     string(userID),
		PlanCode:   plan.Code,
		AmountRUB:  plan.AmountRUB,
		PeriodDays: plan.PeriodDays,
		Currency:   plan.Currency,
		Status:     subscription.OrderPending,
		CreatedAt:  s.clock.Now(),
	}
	if s.orders == nil {
		return nil, apperrors.New(apperrors.CodeInternal, "orders repository not configured")
	}
	if err := s.orders.Create(ctx, order); err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to create order", err)
	}
	return order, nil
}

// CreatePayment starts a new payment for the user's subscription.
// Flow: Order (existing or new) → Payment → Provider URL. Never activates access.
func (s *Service) CreatePayment(ctx context.Context, userID user.ID, planCode string) (confirmationURL string, err error) {
	return s.CreatePaymentForOrder(ctx, userID, planCode, "")
}

// CreatePaymentForOrder creates a provider payment for planCode and/or existing orderID.
func (s *Service) CreatePaymentForOrder(ctx context.Context, userID user.ID, planCode, orderID string) (confirmationURL string, err error) {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return "", err
	}

	var plan Plan
	if orderID != "" {
		if s.orders == nil {
			return "", apperrors.New(apperrors.CodeInternal, "orders repository not configured")
		}
		order, err := s.orders.FindByID(ctx, orderID)
		if err != nil {
			return "", err
		}
		if order.UserID != string(userID) {
			return "", apperrors.New(apperrors.CodeForbidden, "order does not belong to user")
		}
		if order.Status != subscription.OrderPending {
			return "", apperrors.New(apperrors.CodeInvalidInput, "order is not pending").
				WithDetails(map[string]any{"status": string(order.Status)})
		}
		plan, err = s.resolvePlan(order.PlanCode)
		if err != nil {
			return "", err
		}
		// Prefer amounts locked on the order row.
		plan.AmountRUB = order.AmountRUB
		plan.PeriodDays = order.PeriodDays
		if order.Currency != "" {
			plan.Currency = order.Currency
		}
	} else {
		var err error
		plan, err = s.resolvePlan(planCode)
		if err != nil {
			return "", err
		}
		if s.orders != nil {
			order, err := s.CreateOrder(ctx, userID, plan.Code)
			if err != nil {
				return "", err
			}
			orderID = order.ID
		}
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
		payment := &subscription.Payment{
			ID:         paymentID,
			UserID:     string(userID),
			ProviderID: paymentID,
			OrderID:    orderID,
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
			slog.String("order_id", orderID),
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
		OrderID:    orderID,
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
	ok, err := s.payments.MarkSucceededByIDIfPending(ctx, payment.ID, chargeID, tgUserID)
	if err != nil {
		return time.Time{}, err
	}
	if !ok {
		u, err := s.users.FindByID(ctx, user.ID(payment.UserID))
		if err != nil {
			return time.Time{}, err
		}
		if u.SubscriptionActiveUntil != nil {
			return *u.SubscriptionActiveUntil, nil
		}
		return s.clock.Now(), nil
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
	ok, err := s.payments.MarkSucceededIfPending(ctx, payment.ProviderID)
	if err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to mark payment succeeded", err)
	}
	if !ok {
		// Already credited by a concurrent/replayed webhook.
		return nil
	}
	_, err = s.extendFromPayment(ctx, payment)
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
	planCode := payment.Tariff
	if planCode == "" {
		planCode = "personal_basic"
	}
	if err := s.users.ActivatePaidPlan(ctx, user.ID(payment.UserID), newExpiry, planCode); err != nil {
		s.log.Error(ctx, err)
		return time.Time{}, apperrors.Wrap(apperrors.CodeInternal, "failed to extend subscription", err)
	}
	if payment.OrderID != "" && s.orders != nil {
		if _, err := s.orders.MarkPaidIfPending(ctx, payment.OrderID); err != nil {
			s.log.Error(ctx, err)
			// Payment already succeeded; don't fail the webhook on order mark.
		}
	}
	return newExpiry, nil
}

// GetSubscription implements "GET /subscription".
func (s *Service) GetSubscription(ctx context.Context, userID user.ID) (subscription.Info, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return subscription.Info{}, err
	}
	return subscription.DeriveInfo(subscription.EntitlementInput{
		ActiveUntil: u.SubscriptionActiveUntil,
		TrialEndsAt: u.TrialEndsAt,
		Source:      u.EntitlementSource,
		PlanCode:    u.PlanCode,
		CanceledAt:  u.SubscriptionCanceledAt,
		Now:         s.clock.Now(),
	}), nil
}

// ListPayments implements "GET /payments" history (E06).
func (s *Service) ListPayments(ctx context.Context, userID user.ID) ([]*subscription.Payment, error) {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.payments.ListByUserID(ctx, string(userID))
}

// CancelSubscription implements "отмена" (FS E06 / BILLING-001 CANCELED).
func (s *Service) CancelSubscription(ctx context.Context, userID user.ID) error {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return err
	}
	if err := s.users.CancelAutoRenew(ctx, userID, s.clock.Now()); err != nil {
		return err
	}
	s.log.Info(ctx, "subscription cancel recorded (access until active_until)")
	return nil
}

// PaymentsRepo exposes the payment repository for Telegram/USDT modules.
func (s *Service) PaymentsRepo() subscription.PaymentRepository { return s.payments }

// UsersRepo exposes users for Telegram/USDT modules.
func (s *Service) UsersRepo() user.Repository { return s.users }

// ClockNow returns current time.
func (s *Service) ClockNow() time.Time { return s.clock.Now() }
