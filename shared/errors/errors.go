// Package errors defines a single unified application error type used across
// every StreamPass module (backend, client-core). All errors that cross a
// package boundary must be *AppError so that logging, HTTP mapping and
// telemetry can treat errors uniformly.
package errors

import "fmt"

// Code is a stable, machine-readable error identifier. Codes must never be
// reused for a different meaning and must never be removed once shipped,
// since clients and dashboards may match on them.
type Code string

const (
	// Generic
	CodeUnknown       Code = "UNKNOWN"
	CodeInvalidInput  Code = "INVALID_INPUT"
	CodeNotFound      Code = "NOT_FOUND"
	CodeAlreadyExists Code = "ALREADY_EXISTS"
	CodeUnauthorized  Code = "UNAUTHORIZED"
	CodeForbidden     Code = "FORBIDDEN"
	CodeRateLimited   Code = "RATE_LIMITED"
	CodeInternal      Code = "INTERNAL"
	CodeUnavailable   Code = "UNAVAILABLE"
	CodeTimeout       Code = "TIMEOUT"
	CodeConflict      Code = "CONFLICT"

	// Auth
	CodeInvalidCredentials Code = "AUTH_INVALID_CREDENTIALS"
	CodeTokenExpired       Code = "AUTH_TOKEN_EXPIRED"
	CodeTokenInvalid       Code = "AUTH_TOKEN_INVALID"
	CodeTokenRevoked       Code = "AUTH_TOKEN_REVOKED"

	// Billing
	CodeSubscriptionExpired Code = "BILLING_SUBSCRIPTION_EXPIRED"
	CodePaymentFailed       Code = "BILLING_PAYMENT_FAILED"

	// Rule / Config
	CodeRuleSetInvalid Code = "RULE_SET_INVALID"
)

// AppError is the single error type every module returns across package
// boundaries. It intentionally has no interface indirection (KISS) — one
// concrete struct is enough for the whole project.
type AppError struct {
	Code    Code
	Message string
	Details map[string]any
	Cause   error
}

// Error implements the standard error interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap allows errors.Is / errors.As to see through to the cause.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates an AppError without a details map or cause.
func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap creates an AppError that carries an underlying cause.
func Wrap(code Code, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Cause: cause}
}

// WithDetails attaches structured context (e.g. field name, offending
// value) and returns the same error for chaining.
func (e *AppError) WithDetails(details map[string]any) *AppError {
	e.Details = details
	return e
}

// CodeOf extracts the Code from any error, returning CodeUnknown if the
// error is not an *AppError.
func CodeOf(err error) Code {
	if ae, ok := err.(*AppError); ok {
		return ae.Code
	}
	return CodeUnknown
}
