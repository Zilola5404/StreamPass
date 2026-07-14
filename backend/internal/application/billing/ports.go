// Package billing is the Billing application layer (spec section 15):
// creating payments, handling provider webhooks, and reporting
// subscription status. The payment provider itself is abstracted behind
// PaymentProvider so a non-ЮKassa provider can be added later without
// touching use-case logic, per the spec's explicit requirement.
package billing

import (
	"context"
	"time"
)

// PaymentStatus is the provider-agnostic result of checking a payment.
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusSucceeded PaymentStatus = "SUCCEEDED"
	PaymentStatusCanceled  PaymentStatus = "CANCELED"
)

// PaymentProvider is the port every payment provider integration
// implements (ЮKassa today; abstraction is what lets a second provider be
// added later per spec section 15).
type PaymentProvider interface {
	// CreatePayment starts a payment and returns the provider's payment ID
	// plus a URL the client should be redirected to for confirmation.
	CreatePayment(ctx context.Context, userID string, amountRUB int64, description string) (providerPaymentID string, confirmationURL string, err error)
	// FetchPaymentStatus asks the provider directly for a payment's
	// current status. Webhooks are a notification-to-check-again, not a
	// source of truth by themselves — the use case always confirms via
	// this call before crediting a subscription, so a forged/replayed
	// webhook body can't grant access on its own.
	FetchPaymentStatus(ctx context.Context, providerPaymentID string) (PaymentStatus, error)
}

// Clock is injected for deterministic tests.
type Clock interface {
	Now() time.Time
}

// SystemClock is the production Clock implementation.
type SystemClock struct{}

// Now returns the current wall-clock time.
func (SystemClock) Now() time.Time { return time.Now().UTC() }
