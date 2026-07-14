// Package http contains the backend's HTTP transport layer: routing,
// middleware and handlers. Handlers translate between JSON <-> application
// layer calls and never contain business logic themselves (Clean
// Architecture: Interface/Delivery layer).
package http

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrors "streampass/shared/errors"
)

// errorResponse is the single JSON error shape returned by every endpoint,
// so clients can rely on one parsing path regardless of which module
// produced the error.
type errorResponse struct {
	Error struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Details map[string]any `json:"details,omitempty"`
	} `json:"error"`
}

// TimeFormat is the single timestamp format every JSON response uses
// (RFC 3339), so clients parse dates the same way regardless of endpoint.
const TimeFormat = "2006-01-02T15:04:05Z07:00"

// WriteJSON writes v as a JSON body with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, v)
}

// WriteError maps any error to a JSON error body and the appropriate HTTP
// status code (spec: unified error handling).
func WriteError(w http.ResponseWriter, err error) {
	writeError(w, err)
}

// DecodeJSON reads and decodes a JSON request body into dst, returning a
// well-formed AppError on any failure.
func DecodeJSON(r *http.Request, dst any) error {
	return decodeJSON(r, dst)
}

// ErrUnauthenticated builds the AppError a handler returns if it expected
// RequireAuth to have populated the request context but it didn't — this
// should be unreachable in practice (the router always wires RequireAuth
// ahead of these handlers) but the check keeps the handler safe against a
// future routing mistake instead of a nil-pointer panic.
func ErrUnauthenticated() error {
	return apperrors.New(apperrors.CodeUnauthorized, "authentication required")
}

// writeJSON writes v as a JSON body with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError maps any error to a JSON error body and the appropriate HTTP
// status code, based on its AppError code (spec: unified error handling).
func writeError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		appErr = apperrors.Wrap(apperrors.CodeInternal, "internal error", err)
	}

	resp := errorResponse{}
	resp.Error.Code = string(appErr.Code)
	resp.Error.Message = appErr.Message
	resp.Error.Details = appErr.Details

	writeJSON(w, statusForCode(appErr.Code), resp)
}

func statusForCode(code apperrors.Code) int {
	switch code {
	case apperrors.CodeInvalidInput, apperrors.CodeRuleSetInvalid:
		return http.StatusBadRequest
	case apperrors.CodeUnauthorized, apperrors.CodeInvalidCredentials,
		apperrors.CodeTokenExpired, apperrors.CodeTokenInvalid, apperrors.CodeTokenRevoked:
		return http.StatusUnauthorized
	case apperrors.CodeForbidden:
		return http.StatusForbidden
	case apperrors.CodeNotFound:
		return http.StatusNotFound
	case apperrors.CodeAlreadyExists, apperrors.CodeConflict:
		return http.StatusConflict
	case apperrors.CodeRateLimited:
		return http.StatusTooManyRequests
	case apperrors.CodeSubscriptionExpired, apperrors.CodePaymentFailed:
		return http.StatusPaymentRequired
	case apperrors.CodeUnavailable:
		return http.StatusServiceUnavailable
	case apperrors.CodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// decodeJSON reads and decodes a JSON request body into dst, returning a
// well-formed AppError on any failure so the handler doesn't need its own
// error-mapping branch.
func decodeJSON(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return apperrors.Wrap(apperrors.CodeInvalidInput, "malformed request body", err)
	}
	return nil
}
