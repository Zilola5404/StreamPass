// Package subscription contains the Billing bounded context's domain model.
package subscription

import "time"

// Status is the coarse state of a user's subscription / trial (BILLING-002 SSOT).
//
// Official subscription states:
//
//	TRIAL     — free trial window; Connect allowed
//	ACTIVE    — paid (or admin) entitlement; Connect allowed
//	CANCELED  — auto-renew canceled; Connect still allowed until ActiveUntil
//	EXPIRED   — had entitlement, now ended; Connect blocked
//	INACTIVE  — never entitled; Connect blocked
//
// Payment provider statuses (PENDING/SUCCEEDED/CANCELED) are separate and must
// not be confused with these subscription states.
type Status string

const (
	StatusTrial    Status = "TRIAL"
	StatusActive   Status = "ACTIVE"
	StatusExpired  Status = "EXPIRED"
	StatusCanceled Status = "CANCELED"
	StatusInactive Status = "INACTIVE"
)

// Info is the read model returned by "GET /subscription".
type Info struct {
	Status      Status
	ActiveUntil *time.Time
	TrialEndsAt *time.Time
	// Source: trial | paid | admin | ""
	Source string
	// PlanCode last paid/assigned plan (personal_basic | personal_pro | business).
	PlanCode string
	// DaysLeft / HoursLeft remaining access (server clock).
	DaysLeft  int
	HoursLeft int
	// AccessAllowed is true when Connect may start.
	// CANCELED with remaining time still allows Connect until ActiveUntil.
	AccessAllowed bool
	// MaxDevices from plan catalog (0 = use auth.max_devices fallback).
	MaxDevices int
	// ErrorCode for client UX (e.g. TRIAL_EXPIRED).
	ErrorCode string
}

// EntitlementInput is everything needed to derive Info.
type EntitlementInput struct {
	ActiveUntil *time.Time
	TrialEndsAt *time.Time
	Source      string
	PlanCode    string
	CanceledAt  *time.Time
	Now         time.Time
}

// NewInfo derives subscription Info from a raw expiry timestamp.
func NewInfo(activeUntil *time.Time, now time.Time) Info {
	return DeriveInfo(EntitlementInput{ActiveUntil: activeUntil, Now: now})
}

// NewInfoWithTrial builds Info including trial / source fields.
func NewInfoWithTrial(activeUntil, trialEndsAt *time.Time, source string, now time.Time) Info {
	return DeriveInfo(EntitlementInput{
		ActiveUntil: activeUntil,
		TrialEndsAt: trialEndsAt,
		Source:      source,
		Now:         now,
	})
}

// DeriveInfo is the SSOT for subscription status (BILLING-001 / BILLING-002).
func DeriveInfo(in EntitlementInput) Info {
	info := Info{
		ActiveUntil: in.ActiveUntil,
		TrialEndsAt: in.TrialEndsAt,
		Source:      in.Source,
	}
	if in.PlanCode != "" {
		info.PlanCode = NormalizePlanCode(in.PlanCode)
	}
	if n := MaxDevicesForPlan(in.PlanCode); n > 0 {
		info.MaxDevices = n
	} else if in.Source == "trial" || in.PlanCode == "" {
		// Trial / no plan yet → Personal Basic device cap.
		info.MaxDevices = MaxDevicesForPlan(PlanPersonalBasic)
	}
	now := in.Now

	if in.ActiveUntil != nil && in.ActiveUntil.After(now) {
		rem := in.ActiveUntil.Sub(now)
		info.HoursLeft = int(rem.Hours())
		info.DaysLeft = int(rem.Hours() / 24)
		if info.HoursLeft < 0 {
			info.HoursLeft = 0
		}
		if info.DaysLeft < 0 {
			info.DaysLeft = 0
		}
		info.AccessAllowed = true

		if in.CanceledAt != nil {
			info.Status = StatusCanceled
			return info
		}
		if in.Source == "trial" ||
			(in.TrialEndsAt != nil && !in.TrialEndsAt.Before(*in.ActiveUntil) &&
				in.Source != "paid" && in.Source != "admin") {
			info.Status = StatusTrial
			return info
		}
		info.Status = StatusActive
		return info
	}

	info.AccessAllowed = false
	if in.ActiveUntil != nil || in.TrialEndsAt != nil {
		info.Status = StatusExpired
		// TRIAL_EXPIRED only when entitlement was trial (never for paid/admin expiry).
		if in.Source == "trial" {
			info.ErrorCode = "TRIAL_EXPIRED"
		}
		return info
	}
	info.Status = StatusInactive
	return info
}

// Order is a plan purchase intent before (or with) a provider payment.
type Order struct {
	ID         string
	UserID     string
	PlanCode   string
	AmountRUB  int64
	PeriodDays int
	Currency   string
	Status     OrderStatus
	CreatedAt  time.Time
	PaidAt     *time.Time
}

// OrderStatus for orders table.
type OrderStatus string

const (
	OrderPending  OrderStatus = "PENDING"
	OrderPaid     OrderStatus = "PAID"
	OrderCanceled OrderStatus = "CANCELED"
)

// Payment records one payment-provider transaction.
type Payment struct {
	ID             string
	UserID         string
	ProviderID     string
	OrderID        string
	AmountRUB      int64
	PeriodDays     int
	Status         PaymentStatus
	CreatedAt      time.Time
	Provider       string // yookassa | telegram | usdt | platega
	Currency       string
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
