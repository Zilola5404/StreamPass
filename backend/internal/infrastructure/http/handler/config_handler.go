package handler

import (
	"net/http"

	configsvc "streampass/backend/internal/application/configsvc"
	appconfig "streampass/backend/internal/domain/appconfig"
	httpx "streampass/backend/internal/infrastructure/http"
)

// ConfigHandler exposes GET /config and POST /config (admin).
type ConfigHandler struct {
	svc *configsvc.Service
}

// NewConfigHandler builds the Config Service HTTP handler.
func NewConfigHandler(svc *configsvc.Service) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

type configResponse struct {
	Version               int    `json:"version"`
	MinSupportedClientVer string `json:"min_supported_client_version"`
	TelemetryEnabled      bool   `json:"telemetry_enabled"`
	RulePollIntervalSec   int    `json:"rule_poll_interval_sec"`
	RelayPollIntervalSec  int    `json:"relay_poll_interval_sec"`
	UpdatedAt             string `json:"updated_at"`
}

// GetLatest handles "GET /config".
func (h *ConfigHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.svc.GetLatest(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toConfigResponse(cfg))
}

type publishConfigRequest struct {
	MinSupportedClientVer string `json:"min_supported_client_version"`
	TelemetryEnabled      bool   `json:"telemetry_enabled"`
	RulePollIntervalSec   int    `json:"rule_poll_interval_sec"`
	RelayPollIntervalSec  int    `json:"relay_poll_interval_sec"`
}

// Publish handles "POST /config" (admin-only, gated by RequireAdminKey in
// the router).
func (h *ConfigHandler) Publish(w http.ResponseWriter, r *http.Request) {
	var req publishConfigRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	cfg := appconfig.Config{
		MinSupportedClientVer: req.MinSupportedClientVer,
		TelemetryEnabled:      req.TelemetryEnabled,
		RulePollIntervalSec:   req.RulePollIntervalSec,
		RelayPollIntervalSec:  req.RelayPollIntervalSec,
	}

	published, err := h.svc.Publish(r.Context(), cfg)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toConfigResponse(published))
}

func toConfigResponse(cfg *appconfig.Config) configResponse {
	return configResponse{
		Version:               cfg.Version,
		MinSupportedClientVer: cfg.MinSupportedClientVer,
		TelemetryEnabled:      cfg.TelemetryEnabled,
		RulePollIntervalSec:   cfg.RulePollIntervalSec,
		RelayPollIntervalSec:  cfg.RelayPollIntervalSec,
		UpdatedAt:             cfg.UpdatedAt.Format(httpx.TimeFormat),
	}
}
