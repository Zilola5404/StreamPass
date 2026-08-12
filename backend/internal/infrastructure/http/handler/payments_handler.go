package handler

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"

	paymentssvc "streampass/backend/internal/application/payments"
	httpx "streampass/backend/internal/infrastructure/http"
	"streampass/backend/internal/infrastructure/http/middleware"
	tg "streampass/backend/internal/infrastructure/payment/telegram"
	apperrors "streampass/shared/errors"
)

// PaymentsHandler exposes Telegram webhook + USDT endpoints (TZ 02.3 / BL-040).
type PaymentsHandler struct {
	svc           *paymentssvc.Service
	webhookSecret string
}

// NewPaymentsHandler builds the payments HTTP handler.
func NewPaymentsHandler(svc *paymentssvc.Service, webhookSecret string) *PaymentsHandler {
	return &PaymentsHandler{svc: svc, webhookSecret: webhookSecret}
}

// HandleTelegramWebhook handles POST /payments/telegram/webhook.
func (h *PaymentsHandler) HandleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if h.webhookSecret != "" {
		got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if subtle.ConstantTimeCompare([]byte(got), []byte(h.webhookSecret)) != 1 {
			httpx.WriteError(w, apperrors.New(apperrors.CodeForbidden, "invalid webhook secret"))
			return
		}
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, apperrors.New(apperrors.CodeInvalidInput, "failed to read body"))
		return
	}
	var upd tg.Update
	if err := json.Unmarshal(raw, &upd); err != nil {
		httpx.WriteError(w, apperrors.New(apperrors.CodeInvalidInput, "invalid telegram update"))
		return
	}
	if err := h.svc.HandleTelegramUpdate(r.Context(), upd); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type telegramCreateRequest struct {
	Tariff string `json:"tariff"`
}

type telegramCreateResponse struct {
	InvoiceLink string `json:"invoice_link"`
}

// CreateTelegramPayment handles POST /payments/telegram/create (authenticated).
func (h *PaymentsHandler) CreateTelegramPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	var req telegramCreateRequest
	_ = httpx.DecodeJSON(r, &req)
	if req.Tariff == "" {
		req.Tariff = "month"
	}
	link, err := h.svc.CreateTelegramInvoice(r.Context(), userID, req.Tariff)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, telegramCreateResponse{InvoiceLink: link})
}

// USDTAddress handles GET /payments/usdt/address.
func (h *PaymentsHandler) USDTAddress(w http.ResponseWriter, r *http.Request) {
	addr, network, err := h.svc.USDTAddress()
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"address": addr,
		"network": network,
	})
}

type usdtConfirmRequest struct {
	TxHash string `json:"tx_hash"`
	Email  string `json:"email"`
	Tariff string `json:"tariff"`
}

// ConfirmUSDT handles POST /payments/usdt/confirm.
func (h *PaymentsHandler) ConfirmUSDT(w http.ResponseWriter, r *http.Request) {
	var req usdtConfirmRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.svc.ConfirmUSDT(r.Context(), req.TxHash, req.Email, req.Tariff); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}
