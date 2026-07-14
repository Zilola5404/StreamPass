package subscription

import "context"

// PaymentRepository is the port for persisting payment records, used for
// webhook idempotency (a provider may redeliver the same webhook).
type PaymentRepository interface {
	FindByProviderID(ctx context.Context, providerID string) (*Payment, error)
	Create(ctx context.Context, p *Payment) error
	MarkSucceeded(ctx context.Context, providerID string) error
}
