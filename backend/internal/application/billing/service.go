package billing

import (
	"context"
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
	Code       string // "month" | "year"
	Title      string
	AmountRUB  int64
	PeriodDays int
}

// Service implements the Billing use cases.
type Service struct {
	users    user.Repository
	payments subscription.PaymentRepository
	provider PaymentProvider
	plans    []Plan
	clock    Clock
	log      *logger.Logger
}

// NewService wires the Billing service via constructor injection.
func NewService(users user.Repository, payments subscription.PaymentRepository, provider PaymentProvider, plans []Plan, clock Clock, log *logger.Logger) *Service {
	if len(plans) == 0 {
		plans = []Plan{{Code: "month", Title: "Месяц", AmountRUB: 299, PeriodDays: 30}}
	}
	return &Service{users: users, payments: payments, provider: provider, plans: plans, clock: clock, log: log.With("billing_service")}
}

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

// CreatePayment starts a new payment for the user's subscription
// (spec: "подписка", "продление" both go through this — a renewal is just
// another payment that extends ActiveUntil further).
func (s *Service) CreatePayment(ctx context.Context, userID user.ID, planCode string) (confirmationURL string, err error) {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return "", err
	}
	plan, err := s.resolvePlan(planCode)
	if err != nil {
		return "", err
	}

	providerPaymentID, url, err := s.provider.CreatePayment(ctx, string(userID), plan.AmountRUB, "StreamPass "+plan.Title)
	if err != nil {
		s.log.Error(ctx, err)
		return "", apperrors.Wrap(apperrors.CodePaymentFailed, "failed to create payment", err)
	}

	payment := &subscription.Payment{
		ID:         idgen.New(),
		UserID:     string(userID),
		ProviderID: providerPaymentID,
		AmountRUB:  plan.AmountRUB,
		PeriodDays: plan.PeriodDays,
		Status:     subscription.PaymentPending,
		CreatedAt:  s.clock.Now(),
	}
	if err := s.payments.Create(ctx, payment); err != nil {
		s.log.Error(ctx, err)
		return "", apperrors.Wrap(apperrors.CodeInternal, "failed to record pending payment", err)
	}

	return url, nil
}

// HandleWebhook processes a payment-provider notification. It never trusts
// the webhook body's claimed status directly — it re-fetches status from
// the provider (see PaymentProvider.FetchPaymentStatus doc) before
// crediting a subscription, and is idempotent against redelivery.
func (s *Service) HandleWebhook(ctx context.Context, providerPaymentID string) error {
	payment, err := s.payments.FindByProviderID(ctx, providerPaymentID)
	if err != nil {
		return err
	}
	if payment.Status == subscription.PaymentSucceeded {
		return nil // already processed; webhook redelivery is a no-op
	}

	status, err := s.provider.FetchPaymentStatus(ctx, providerPaymentID)
	if err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodePaymentFailed, "failed to confirm payment status with provider", err)
	}
	if status != PaymentStatusSucceeded {
		return nil // still pending or canceled — nothing to credit yet
	}

	newExpiry := s.clock.Now().Add(time.Duration(payment.PeriodDays) * 24 * time.Hour)
	if err := s.users.ExtendSubscription(ctx, user.ID(payment.UserID), newExpiry); err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to extend subscription", err)
	}
	if err := s.payments.MarkSucceeded(ctx, providerPaymentID); err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to mark payment succeeded", err)
	}
	return nil
}

// GetSubscription implements "GET /subscription".
func (s *Service) GetSubscription(ctx context.Context, userID user.ID) (subscription.Info, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return subscription.Info{}, err
	}
	return subscription.NewInfo(u.SubscriptionActiveUntil, s.clock.Now()), nil
}

// ListPayments implements "GET /payments" history (E06).
func (s *Service) ListPayments(ctx context.Context, userID user.ID) ([]*subscription.Payment, error) {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.payments.ListByUserID(ctx, string(userID))
}

// CancelSubscription implements "отмена" (FS E06): access remains until
// active_until. Without auto-renewal (BL-030), cancel is an acknowledge
// no-op on the server — the client stops offering renewal UX.
func (s *Service) CancelSubscription(ctx context.Context, userID user.ID) error {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return err
	}
	s.log.Info(ctx, "subscription cancel acknowledged (access until active_until)")
	return nil
}
