package handler

import (
	"net/http"
	"time"

	relaysvc "streampass/backend/internal/application/relay"
	relaydomain "streampass/backend/internal/domain/relay"
	httpx "streampass/backend/internal/infrastructure/http"
)

// RelayHandler exposes GET /servers and POST /servers/health.
type RelayHandler struct {
	svc *relaysvc.Service
}

// NewRelayHandler builds the Relay Manager HTTP handler.
func NewRelayHandler(svc *relaysvc.Service) *RelayHandler {
	return &RelayHandler{svc: svc}
}

type serverDTO struct {
	ID        string  `json:"id"`
	Region    string  `json:"region"`
	Host      string  `json:"host"`
	Port      int     `json:"port"`
	LoadRatio float64 `json:"load_ratio"`
	RTTMillis int     `json:"rtt_ms"`
}

// ListAvailable handles "GET /servers".
func (h *RelayHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	servers, err := h.svc.ListAvailable(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	dtos := make([]serverDTO, len(servers))
	for i, s := range servers {
		dtos[i] = serverDTO{
			ID:        string(s.ID),
			Region:    string(s.Region),
			Host:      s.Host,
			Port:      s.Port,
			LoadRatio: s.LoadRatio,
			RTTMillis: s.RTTMillis,
		}
	}
	httpx.WriteJSON(w, http.StatusOK, dtos)
}

type healthCheckRequest struct {
	ID        string  `json:"id"`
	Healthy   bool    `json:"healthy"`
	LoadRatio float64 `json:"load_ratio"`
	RTTMillis int     `json:"rtt_ms"`
}

// RecordHealthCheck handles "POST /servers/health" — called by the Health
// Monitor component, gated by RequireAdminKey in the router since it's an
// internal-only write path, not a client-facing one.
func (h *RelayHandler) RecordHealthCheck(w http.ResponseWriter, r *http.Request) {
	var req healthCheckRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	err := h.svc.RecordHealthCheck(r.Context(), relaydomain.ID(req.ID), req.Healthy, req.LoadRatio, req.RTTMillis, time.Now().UTC())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}
