package subscription

import "context"

// PaymentRepository is the port for persisting payment records, used for
// webhook idempotency (a provider may redeliver the same webhook).
type PaymentRepository interface {
	FindByProviderID(ctx context.Context, providerID string) (*Payment, error)
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByTxHash(ctx context.Context, txHash string) (*Payment, error)
	Create(ctx context.Context, p *Payment) error
	// MarkSucceededIfPending returns true only when this call transitioned PENDING→SUCCEEDED.
	// Concurrent/replayed webhooks return false without error (no duplicate subscription extend).
	MarkSucceededIfPending(ctx context.Context, providerID string) (bool, error)
	MarkSucceededByIDIfPending(ctx context.Context, id, chargeID string, tgUserID *int64) (bool, error)
	ListByUserID(ctx context.Context, userID string) ([]*Payment, error)
}

// OrderRepository persists purchase intents (BILLING-001 Create Order step).
type OrderRepository interface {
	Create(ctx context.Context, o *Order) error
	FindByID(ctx context.Context, id string) (*Order, error)
	MarkPaidIfPending(ctx context.Context, id string) (bool, error)
}
