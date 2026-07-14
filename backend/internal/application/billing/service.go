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

// Plan describes the single subscription plan the MVP sells. Loaded from
// config (spec: "Запрещено использовать hardcode"), not a constant, so
// price/period can change without a code deploy.
type Plan struct {
	AmountRUB  int64
	PeriodDays int
}

// Service implements the Billing use cases.
type Service struct {
	users    user.Repository
	payments subscription.PaymentRepository
	provider PaymentProvider
	plan     Plan
	clock    Clock
	log      *logger.Logger
}

// NewService wires the Billing service via constructor injection.
func NewService(users user.Repository, payments subscription.PaymentRepository, provider PaymentProvider, plan Plan, clock Clock, log *logger.Logger) *Service {
	return &Service{users: users, payments: payments, provider: provider, plan: plan, clock: clock, log: log.With("billing_service")}
}

// CreatePayment starts a new payment for the user's subscription
// (spec: "подписка", "продление" both go through this — a renewal is just
// another payment that extends ActiveUntil further).
func (s *Service) CreatePayment(ctx context.Context, userID user.ID) (confirmationURL string, err error) {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return "", err
	}

	providerPaymentID, url, err := s.provider.CreatePayment(ctx, string(userID), s.plan.AmountRUB, "StreamPass subscription")
	if err != nil {
		s.log.Error(ctx, err)
		return "", apperrors.Wrap(apperrors.CodePaymentFailed, "failed to create payment", err)
	}

	payment := &subscription.Payment{
		ID:         idgen.New(),
		UserID:     string(userID),
		ProviderID: providerPaymentID,
		AmountRUB:  s.plan.AmountRUB,
		PeriodDays: s.plan.PeriodDays,
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

// CancelSubscription implements "отмена" (spec section 15): immediately
// revokes access. The spec does not describe recurring/auto-renewal
// billing mechanics, so "cancel" here means "end access now" rather than
// "stop a future auto-charge" — flagged in project recommendations as a
// point needing product clarification.
func (s *Service) CancelSubscription(ctx context.Context, userID user.ID) error {
	if err := s.users.ExtendSubscription(ctx, userID, s.clock.Now()); err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to cancel subscription", err)
	}
	return nil
}
