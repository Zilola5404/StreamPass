package handler

import (
	"crypto/subtle"
	"net/http"
	"strings"

	billingsvc "streampass/backend/internal/application/billing"
	httpx "streampass/backend/internal/infrastructure/http"
	"streampass/backend/internal/infrastructure/http/middleware"
	apperrors "streampass/shared/errors"
)

// BillingHandler exposes POST /payments, GET /payments, GET /plans,
// POST /payments/webhook, GET /subscription, POST /subscription/cancel.
type BillingHandler struct {
	svc           *billingsvc.Service
	webhookSecret string
	// plategaAuth verifies X-MerchantId / X-Secret on Platega callbacks (optional).
	plategaAuth func(merchantID, secret string) bool
}

// NewBillingHandler builds the Billing HTTP handler.
// webhookSecret is optional; when set, callers must send header
// X-StreamPass-Webhook-Secret (defense in depth — S-04).
func NewBillingHandler(svc *billingsvc.Service, webhookSecret string) *BillingHandler {
	return &BillingHandler{svc: svc, webhookSecret: webhookSecret}
}

// SetPlategaAuth enables Platega callback header verification.
func (h *BillingHandler) SetPlategaAuth(fn func(merchantID, secret string) bool) {
	h.plategaAuth = fn
}

type createPaymentRequest struct {
	PlanCode string `json:"plan_code"`
	OrderID  string `json:"order_id"`
}

type createPaymentResponse struct {
	ConfirmationURL string `json:"confirmation_url"`
	OrderID         string `json:"order_id,omitempty"`
}

type createOrderRequest struct {
	PlanCode string `json:"plan_code"`
}

type createOrderResponse struct {
	OrderID    string `json:"order_id"`
	PlanCode   string `json:"plan_code"`
	AmountRUB  int64  `json:"amount_rub"`
	PeriodDays int    `json:"period_days"`
	Currency   string `json:"currency"`
	Status     string `json:"status"`
}

// CreateOrder handles "POST /orders" (authenticated) — BILLING-001.
func (h *BillingHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	var req createOrderRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	order, err := h.svc.CreateOrder(r.Context(), userID, req.PlanCode)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, createOrderResponse{
		OrderID:    order.ID,
		PlanCode:   order.PlanCode,
		AmountRUB:  order.AmountRUB,
		PeriodDays: order.PeriodDays,
		Currency:   order.Currency,
		Status:     string(order.Status),
	})
}

// CreatePayment handles "POST /payments" (authenticated).
func (h *BillingHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}

	var req createPaymentRequest
	// Empty body is fine — defaults to first plan.
	_ = httpx.DecodeJSON(r, &req)

	planCode := req.PlanCode
	url, err := h.svc.CreatePaymentForOrder(r.Context(), userID, planCode, req.OrderID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, createPaymentResponse{ConfirmationURL: url, OrderID: req.OrderID})
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
	if h.webhookSecret != "" {
		got := r.Header.Get("X-StreamPass-Webhook-Secret")
		if subtle.ConstantTimeCompare([]byte(got), []byte(h.webhookSecret)) != 1 {
			httpx.WriteError(w, apperrors.New(apperrors.CodeForbidden, "invalid webhook secret"))
			return
		}
	}

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

type plategaCallback struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"`
	Payload string `json:"payload"`
}

// HandlePlategaWebhook handles POST /payments/platega/webhook.
// Activates subscription only on CONFIRMED after FetchPaymentStatus confirmation.
func (h *BillingHandler) HandlePlategaWebhook(w http.ResponseWriter, r *http.Request) {
	if h.plategaAuth != nil {
		mid := r.Header.Get("X-MerchantId")
		sec := r.Header.Get("X-Secret")
		if !h.plategaAuth(mid, sec) {
			httpx.WriteError(w, apperrors.New(apperrors.CodeForbidden, "invalid platega credentials"))
			return
		}
	}
	var cb plategaCallback
	if err := httpx.DecodeJSON(r, &cb); err != nil {
		httpx.WriteError(w, err)
		return
	}
	status := strings.ToUpper(strings.TrimSpace(cb.Status))
	if status != "CONFIRMED" {
		// Acknowledge CANCELED / CHARGEBACKED without activating.
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"ok": "ignored"})
		return
	}
	if cb.ID == "" {
		httpx.WriteError(w, apperrors.New(apperrors.CodeInvalidInput, "missing transaction id"))
		return
	}
	if err := h.svc.HandleWebhook(r.Context(), cb.ID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"ok": "confirmed"})
}

type subscriptionResponse struct {
	Status        string  `json:"status"`
	ActiveUntil   *string `json:"active_until,omitempty"`
	TrialEndsAt   *string `json:"trial_ends_at,omitempty"`
	Source        string  `json:"source,omitempty"`
	PlanCode      string  `json:"plan_code,omitempty"`
	DaysLeft      int     `json:"days_left"`
	HoursLeft     int     `json:"hours_left"`
	AccessAllowed bool    `json:"access_allowed"`
	MaxDevices    int     `json:"max_devices,omitempty"`
	ErrorCode     string  `json:"error_code,omitempty"`
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

	resp := subscriptionResponse{
		Status:        string(info.Status),
		Source:        info.Source,
		PlanCode:      info.PlanCode,
		DaysLeft:      info.DaysLeft,
		HoursLeft:     info.HoursLeft,
		AccessAllowed: info.AccessAllowed,
		MaxDevices:    info.MaxDevices,
		ErrorCode:     info.ErrorCode,
	}
	if info.ActiveUntil != nil {
		formatted := info.ActiveUntil.Format(httpx.TimeFormat)
		resp.ActiveUntil = &formatted
	}
	if info.TrialEndsAt != nil {
		formatted := info.TrialEndsAt.Format(httpx.TimeFormat)
		resp.TrialEndsAt = &formatted
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

type planDTO struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	AmountRUB   int64  `json:"amount_rub"`
	PeriodDays  int    `json:"period_days"`
	Currency    string `json:"currency"`
	MaxDevices  int    `json:"max_devices,omitempty"`
	MaxUsers    int    `json:"max_users,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListPlans handles "GET /plans" (authenticated — tariffs for E06 / BILLING-001).
func (h *BillingHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans := h.svc.ListPlans()
	out := make([]planDTO, 0, len(plans))
	for _, p := range plans {
		cur := p.Currency
		if cur == "" {
			cur = "RUB"
		}
		out = append(out, planDTO{
			Code:        p.Code,
			Title:       p.Title,
			AmountRUB:   p.AmountRUB,
			PeriodDays:  p.PeriodDays,
			Currency:    cur,
			MaxDevices:  p.MaxDevices,
			MaxUsers:    p.MaxUsers,
			Description: p.Description,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

type paymentDTO struct {
	ID         string `json:"id"`
	AmountRUB  int64  `json:"amount_rub"`
	PeriodDays int    `json:"period_days"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// ListPayments handles "GET /payments" (authenticated) — payment history.
func (h *BillingHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	list, err := h.svc.ListPayments(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	out := make([]paymentDTO, 0, len(list))
	for _, p := range list {
		out = append(out, paymentDTO{
			ID:         p.ID,
			AmountRUB:  p.AmountRUB,
			PeriodDays: p.PeriodDays,
			Status:     string(p.Status),
			CreatedAt:  p.CreatedAt.Format(httpx.TimeFormat),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}
