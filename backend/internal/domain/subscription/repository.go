package subscription

import (
	"context"
	"time"
)

// PaymentRepository is the port for persisting payment records, used for
// webhook idempotency (a provider may redeliver the same webhook).
type PaymentRepository interface {
	FindByProviderID(ctx context.Context, providerID string) (*Payment, error)
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByTxHash(ctx context.Context, txHash string) (*Payment, error)
	Create(ctx context.Context, p *Payment) error
	// Credit atomically settles a pending payment, extends access and updates its order.
	Credit(ctx context.Context, id, chargeID string, tgUserID *int64, now time.Time) (time.Time, bool, error)
	ListByUserID(ctx context.Context, userID string) ([]*Payment, error)
}

// OrderRepository persists purchase intents (BILLING-001 Create Order step).
type OrderRepository interface {
	Create(ctx context.Context, o *Order) error
	FindByID(ctx context.Context, id string) (*Order, error)
	ClaimCheckout(ctx context.Context, id string) (url string, claimed bool, err error)
	SaveCheckoutURL(ctx context.Context, id, url string) error
	MarkPaidIfPending(ctx context.Context, id string) (bool, error)
}
