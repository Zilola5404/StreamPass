package handler

import (
	"net/http"
	"strconv"
	"time"

	diagsvc "streampass/backend/internal/application/diag"
	"streampass/backend/internal/domain/diag"
	httpx "streampass/backend/internal/infrastructure/http"
	"streampass/backend/internal/infrastructure/http/middleware"
)

// DiagHandler exposes POST /diag (client) and GET /admin/diag (operator).
type DiagHandler struct {
	svc *diagsvc.Service
}

// NewDiagHandler builds the diagnostic HTTP handler.
func NewDiagHandler(svc *diagsvc.Service) *DiagHandler {
	return &DiagHandler{svc: svc}
}

type diagEventDTO struct {
	UserID        string `json:"user_id,omitempty"`
	Proto         string `json:"proto"`
	Host          string `json:"host"`
	DestIP        string `json:"dest_ip"`
	DestPort      int    `json:"dest_port"`
	Mode          string `json:"mode"`
	Result        string `json:"result"`
	LatencyMS     int    `json:"latency_ms"`
	ErrorCode     string `json:"error_code"`
	RelayID       string `json:"relay_id"`
	ClientVersion string `json:"client_version"`
	RecordedAt    string `json:"recorded_at,omitempty"`
}

type diagBatchRequest struct {
	Events []diagEventDTO `json:"events"`
}

// RecordBatch handles "POST /diag" (authenticated).
func (h *DiagHandler) RecordBatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	var req diagBatchRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	events := make([]diag.Event, 0, len(req.Events))
	for _, d := range req.Events {
		e := diag.Event{
			Proto:         d.Proto,
			Host:          d.Host,
			DestIP:        d.DestIP,
			DestPort:      d.DestPort,
			Mode:          d.Mode,
			Result:        d.Result,
			LatencyMS:     d.LatencyMS,
			ErrorCode:     d.ErrorCode,
			RelayID:       d.RelayID,
			ClientVersion: d.ClientVersion,
		}
		if d.RecordedAt != "" {
			if t, err := time.Parse(time.RFC3339, d.RecordedAt); err == nil {
				e.RecordedAt = t.UTC()
			}
		}
		events = append(events, e)
	}
	if err := h.svc.RecordBatch(r.Context(), string(userID), events); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

// ListAdmin handles "GET /admin/diag" (admin key).
func (h *DiagHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	list, err := h.svc.List(r.Context(), userID, limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	out := make([]diagEventDTO, 0, len(list))
	for _, e := range list {
		out = append(out, diagEventDTO{
			UserID:        e.UserID,
			Proto:         e.Proto,
			Host:          e.Host,
			DestIP:        e.DestIP,
			DestPort:      e.DestPort,
			Mode:          e.Mode,
			Result:        e.Result,
			LatencyMS:     e.LatencyMS,
			ErrorCode:     e.ErrorCode,
			RelayID:       e.RelayID,
			ClientVersion: e.ClientVersion,
			RecordedAt:    e.RecordedAt.Format(httpx.TimeFormat),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}
