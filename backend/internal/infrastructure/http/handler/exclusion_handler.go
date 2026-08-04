package handler

import (
	"net/http"

	exclusionsvc "streampass/backend/internal/application/exclusion"
	httpx "streampass/backend/internal/infrastructure/http"
	"streampass/backend/internal/infrastructure/http/middleware"
)

// ExclusionHandler exposes GET/PUT /exclusions (BL-014).
type ExclusionHandler struct {
	svc *exclusionsvc.Service
}

// NewExclusionHandler builds the exclusions HTTP handler.
func NewExclusionHandler(svc *exclusionsvc.Service) *ExclusionHandler {
	return &ExclusionHandler{svc: svc}
}

type exclusionsBody struct {
	Domains []string `json:"domains"`
}

// List handles GET /exclusions.
func (h *ExclusionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	domains, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, exclusionsBody{Domains: domains})
}

// Replace handles PUT /exclusions (full list replace).
func (h *ExclusionHandler) Replace(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	var req exclusionsBody
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	domains, err := h.svc.Replace(r.Context(), userID, req.Domains)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, exclusionsBody{Domains: domains})
}
