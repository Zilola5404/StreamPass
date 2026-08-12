package subscription

import "context"

// PaymentRepository is the port for persisting payment records, used for
// webhook idempotency (a provider may redeliver the same webhook).
type PaymentRepository interface {
	FindByProviderID(ctx context.Context, providerID string) (*Payment, error)
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByTxHash(ctx context.Context, txHash string) (*Payment, error)
	Create(ctx context.Context, p *Payment) error
	MarkSucceeded(ctx context.Context, providerID string) error
	MarkSucceededByID(ctx context.Context, id, chargeID string, tgUserID *int64) error
	// ListByUserID returns payments for a user, newest first (E06 history).
	ListByUserID(ctx context.Context, userID string) ([]*Payment, error)
}
