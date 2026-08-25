// Package subscription contains the Billing bounded context's domain
// model. Subscription state itself is stored as a denormalized field on
// user.User (SubscriptionActiveUntil) rather than a separate aggregate,
// since the MVP only needs "is the user currently paid up" — a second
// aggregate with its own repository would be premature (YAGNI). Payment
// history/provider-specific records are what actually live here.
package subscription

import "time"

// Status is the coarse state of a user's subscription / trial.
type Status string

const (
	StatusTrial    Status = "TRIAL"
	StatusActive   Status = "ACTIVE"
	StatusExpired  Status = "EXPIRED"
	StatusInactive Status = "INACTIVE"
)

// Info is the read model returned by "GET /subscription".
type Info struct {
	Status      Status
	ActiveUntil *time.Time
	// TrialEndsAt is set when the user is (or was) on a free trial.
	TrialEndsAt *time.Time
	// Source: trial | paid | admin | ""
	Source string
	// DaysLeft is remaining whole days of access (0 if expired).
	DaysLeft int
}

// NewInfo derives subscription Info from expiry + optional trial metadata.
func NewInfo(activeUntil *time.Time, now time.Time) Info {
	return NewInfoWithTrial(activeUntil, nil, "", now)
}

// NewInfoWithTrial builds Info including trial / source fields.
func NewInfoWithTrial(activeUntil, trialEndsAt *time.Time, source string, now time.Time) Info {
	info := Info{
		ActiveUntil: activeUntil,
		TrialEndsAt: trialEndsAt,
		Source:      source,
	}
	if activeUntil != nil && activeUntil.After(now) {
		info.DaysLeft = int(activeUntil.Sub(now).Hours() / 24)
		if info.DaysLeft < 0 {
			info.DaysLeft = 0
		}
		if source == "trial" || (trialEndsAt != nil && !trialEndsAt.Before(*activeUntil) && source != "paid" && source != "admin") {
			info.Status = StatusTrial
			return info
		}
		info.Status = StatusActive
		return info
	}
	if activeUntil != nil || trialEndsAt != nil {
		info.Status = StatusExpired
		return info
	}
	info.Status = StatusInactive
	return info
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
	Provider       string // yookassa | telegram | usdt | platega
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
