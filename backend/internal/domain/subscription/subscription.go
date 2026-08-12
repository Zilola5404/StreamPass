// Package subscription contains the Billing bounded context's domain
// model. Subscription state itself is stored as a denormalized field on
// user.User (SubscriptionActiveUntil) rather than a separate aggregate,
// since the MVP only needs "is the user currently paid up" — a second
// aggregate with its own repository would be premature (YAGNI). Payment
// history/provider-specific records are what actually live here.
package subscription

import "time"

// Status is the coarse state of a user's subscription.
type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
)

// Info is the read model returned by "GET /subscription".
type Info struct {
	Status      Status
	ActiveUntil *time.Time
}

// NewInfo derives subscription Info from a raw expiry timestamp.
func NewInfo(activeUntil *time.Time, now time.Time) Info {
	if activeUntil != nil && activeUntil.After(now) {
		return Info{Status: StatusActive, ActiveUntil: activeUntil}
	}
	return Info{Status: StatusInactive, ActiveUntil: activeUntil}
}

// Payment records one payment-provider transaction, used for
// audit/idempotency when handling webhooks.
type Payment struct {
	ID             string
	UserID         string
	ProviderID     string // payment ID assigned by the payment provider (or our id for pending Stars)
	AmountRUB      int64  // rubles OR Stars amount when Currency=XTR
	PeriodDays     int
	Status         PaymentStatus
	CreatedAt      time.Time
	Provider       string // yookassa | telegram | usdt
	Currency       string // RUB | XTR | USDT
	TelegramUserID *int64
	Tariff         string
	TxHash         string
	PaidAt         *time.Time
}

// PaymentStatus mirrors the provider-agnostic payment lifecycle.
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "PENDING"
	PaymentSucceeded PaymentStatus = "SUCCEEDED"
	PaymentFailed    PaymentStatus = "FAILED"
)
