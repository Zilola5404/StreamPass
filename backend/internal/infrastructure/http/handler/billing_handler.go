package handler

import (
	"net/http"

	billingsvc "streampass/backend/internal/application/billing"
	httpx "streampass/backend/internal/infrastructure/http"
	"streampass/backend/internal/infrastructure/http/middleware"
)

// BillingHandler exposes POST /payments, POST /payments/webhook,
// GET /subscription, POST /subscription/cancel.
type BillingHandler struct {
	svc *billingsvc.Service
}

// NewBillingHandler builds the Billing HTTP handler.
func NewBillingHandler(svc *billingsvc.Service) *BillingHandler {
	return &BillingHandler{svc: svc}
}

type createPaymentResponse struct {
	ConfirmationURL string `json:"confirmation_url"`
}

// CreatePayment handles "POST /payments" (authenticated).
func (h *BillingHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}

	url, err := h.svc.CreatePayment(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, createPaymentResponse{ConfirmationURL: url})
}

type webhookRequest struct {
	// ProviderPaymentID is read from a top-level field so this handler
	// stays agnostic to which provider sent the webhook — provider-specific
	// payload parsing (ЮKassa's actual notification shape) belongs in the
	// infrastructure/payment/yookassa adapter that constructs this from the
	// raw request before calling HandleWebhook, per the abstraction in
	// billing.PaymentProvider.
	ProviderPaymentID string `json:"provider_payment_id"`
}

// HandleWebhook handles "POST /payments/webhook" (no auth — payment
// providers call this directly; billingsvc.HandleWebhook never trusts the
// body's claimed status and re-fetches it from the provider, so an
// unauthenticated caller can at most trigger a redundant status check).
func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var req webhookRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err := h.svc.HandleWebhook(r.Context(), req.ProviderPaymentID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

type subscriptionResponse struct {
	Status      string  `json:"status"`
	ActiveUntil *string `json:"active_until,omitempty"`
}

// GetSubscription handles "GET /subscription" (authenticated).
func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}

	info, err := h.svc.GetSubscription(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	resp := subscriptionResponse{Status: string(info.Status)}
	if info.ActiveUntil != nil {
		formatted := info.ActiveUntil.Format(httpx.TimeFormat)
		resp.ActiveUntil = &formatted
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// CancelSubscription handles "POST /subscription/cancel" (authenticated).
func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}

	if err := h.svc.CancelSubscription(r.Context(), userID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}
