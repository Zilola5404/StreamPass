package handler

import (
	"net/http"

	telemetrysvc "streampass/backend/internal/application/telemetry"
	telemetrydomain "streampass/backend/internal/domain/telemetry"
	httpx "streampass/backend/internal/infrastructure/http"
	"streampass/backend/internal/infrastructure/http/middleware"
)

// TelemetryHandler exposes POST /telemetry.
type TelemetryHandler struct {
	svc *telemetrysvc.Service
}

// NewTelemetryHandler builds the Telemetry HTTP handler.
func NewTelemetryHandler(svc *telemetrysvc.Service) *TelemetryHandler {
	return &TelemetryHandler{svc: svc}
}

type telemetryEventRequest struct {
	RTTMillis     int     `json:"rtt_ms"`
	PacketLossPct float64 `json:"packet_loss_pct"`
	RelayID       string  `json:"relay_id"`
	ClientVersion string  `json:"client_version"`
	OS            string  `json:"os"`
	ConnectMillis int     `json:"connect_ms"`
	ErrorCode     string  `json:"error_code,omitempty"`
}

// Record handles "POST /telemetry". Requires authentication (via
// RequireAuth in the router) so events are attributable to a user ID
// without the client needing to submit one — and so an unauthenticated
// caller can't flood the events table.
func (h *TelemetryHandler) Record(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}

	var req telemetryEventRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	event := telemetrydomain.Event{
		UserID:        string(userID),
		RTTMillis:     req.RTTMillis,
		PacketLossPct: req.PacketLossPct,
		RelayID:       req.RelayID,
		ClientVersion: req.ClientVersion,
		OS:            req.OS,
		ConnectMillis: req.ConnectMillis,
		ErrorCode:     req.ErrorCode,
	}

	if err := h.svc.Record(r.Context(), event); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}
