package handler

import (
	"net/http"
	"time"

	relaysvc "streampass/backend/internal/application/relay"
	relaydomain "streampass/backend/internal/domain/relay"
	httpx "streampass/backend/internal/infrastructure/http"
	"streampass/shared/region"
)

type RelayHandler struct {
	svc *relaysvc.Service
}

func NewRelayHandler(svc *relaysvc.Service) *RelayHandler {
	return &RelayHandler{svc: svc}
}

type serverDTO struct {
	ID               string  `json:"id"`
	Region           string  `json:"region"`
	RegionName       string  `json:"region_name"`
	Host             string  `json:"host"`
	Port             int     `json:"port"`
	Healthy          bool    `json:"healthy"`
	LoadRatio        float64 `json:"load_ratio"`
	RTTMillis        int     `json:"rtt_ms"`
	ConnectionConfig string  `json:"connection_config"`
}

func toServerDTO(s relaydomain.Server) serverDTO {
	code := region.Normalize(string(s.Region))
	return serverDTO{
		ID:               string(s.ID),
		Region:           code,
		RegionName:       region.LabelOf(code),
		Host:             s.Host,
		Port:             s.Port,
		Healthy:          s.Healthy,
		LoadRatio:        s.LoadRatio,
		RTTMillis:        s.RTTMillis,
		ConnectionConfig: s.ConnectionConfig,
	}
}

func (h *RelayHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("region")
	servers, err := h.svc.ListAvailable(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	dtos := make([]serverDTO, len(servers))
	for i, s := range servers {
		dtos[i] = toServerDTO(s)
	}
	httpx.WriteJSON(w, http.StatusOK, dtos)
}

// ListRegions returns the canonical region catalog (ТЗ Этап 6).
func (h *RelayHandler) ListRegions(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, region.Catalog)
}

// ListAll handles "GET /servers/all" (admin-only, gated by
// RequireAdminKey in the router): returns every registered relay,
// healthy or not. Used by the Health Monitor to find servers that need
// checking, including ones currently marked unhealthy that may have
// recovered.
func (h *RelayHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	servers, err := h.svc.ListAll(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	dtos := make([]serverDTO, len(servers))
	for i, s := range servers {
		dtos[i] = toServerDTO(s)
	}
	httpx.WriteJSON(w, http.StatusOK, dtos)
}

type registerServerRequest struct {
	ID               string `json:"id"`
	Region           string `json:"region"`
	Host             string `json:"host"`
	Port             int    `json:"port"`
	ConnectionConfig string `json:"connection_config"`
}

func (h *RelayHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerServerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	srv, err := h.svc.Register(r.Context(), relaydomain.ID(req.ID), relaydomain.Region(req.Region), req.Host, req.Port, req.ConnectionConfig, time.Now().UTC())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toServerDTO(*srv))
}

func (h *RelayHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.WriteError(w, httpx.ErrMissingPathValue("id"))
		return
	}

	if err := h.svc.Delete(r.Context(), relaydomain.ID(id)); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

type healthCheckRequest struct {
	ID        string  `json:"id"`
	Healthy   bool    `json:"healthy"`
	LoadRatio float64 `json:"load_ratio"`
	RTTMillis int     `json:"rtt_ms"`
}

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
