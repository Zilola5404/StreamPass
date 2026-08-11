package handler

import (
	"net/http"
	"strconv"

	adminsvc "streampass/backend/internal/application/admin"
	httpx "streampass/backend/internal/infrastructure/http"
)

type AdminHandler struct {
	svc *adminsvc.UserService
}

func NewAdminHandler(svc *adminsvc.UserService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

type userSummaryDTO struct {
	ID                      string  `json:"id"`
	Email                   string  `json:"email"`
	CreatedAt               string  `json:"created_at"`
	SubscriptionActiveUntil *string `json:"subscription_active_until,omitempty"`
	SubscriptionActive      bool    `json:"subscription_active"`
	Banned                  bool    `json:"banned"`
	BannedAt                *string `json:"banned_at,omitempty"`
}

func toUserDTO(u adminsvc.UserSummary) userSummaryDTO {
	dto := userSummaryDTO{
		ID:                 string(u.ID),
		Email:              u.Email,
		CreatedAt:          u.CreatedAt.Format(httpx.TimeFormat),
		SubscriptionActive: u.IsSubscriptionActive,
		Banned:             u.Banned,
	}
	if u.SubscriptionActiveUntil != nil {
		formatted := u.SubscriptionActiveUntil.Format(httpx.TimeFormat)
		dto.SubscriptionActiveUntil = &formatted
	}
	if u.BannedAt != nil {
		formatted := u.BannedAt.Format(httpx.TimeFormat)
		dto.BannedAt = &formatted
	}
	return dto
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	users, err := h.svc.ListUsers(r.Context(), q)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	dtos := make([]userSummaryDTO, len(users))
	for i, u := range users {
		dtos[i] = toUserDTO(u)
	}
	httpx.WriteJSON(w, http.StatusOK, dtos)
}

type grantPremiumRequest struct {
	Days int `json:"days"`
}

func adminActor(r *http.Request) string {
	if v := r.Header.Get("X-Admin-Actor"); v != "" {
		return v
	}
	return "admin"
}

// GrantPremium handles POST /users/{id}/subscription.
func (h *AdminHandler) GrantPremium(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.WriteError(w, httpx.ErrMissingPathValue("id"))
		return
	}
	var req grantPremiumRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if req.Days == 0 {
		req.Days = 30
	}
	u, err := h.svc.GrantPremium(r.Context(), adminActor(r), id, req.Days)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserDTO(*u))
}

// RevokePremium handles DELETE /users/{id}/subscription.
func (h *AdminHandler) RevokePremium(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.WriteError(w, httpx.ErrMissingPathValue("id"))
		return
	}
	u, err := h.svc.RevokePremium(r.Context(), adminActor(r), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserDTO(*u))
}

// BanUser handles POST /users/{id}/ban.
func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.WriteError(w, httpx.ErrMissingPathValue("id"))
		return
	}
	u, err := h.svc.BanUser(r.Context(), adminActor(r), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserDTO(*u))
}

// UnbanUser handles DELETE /users/{id}/ban.
func (h *AdminHandler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.WriteError(w, httpx.ErrMissingPathValue("id"))
		return
	}
	u, err := h.svc.UnbanUser(r.Context(), adminActor(r), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserDTO(*u))
}

type auditDTO struct {
	ID         string         `json:"id"`
	Actor      string         `json:"actor"`
	Action     string         `json:"action"`
	TargetType string         `json:"target_type"`
	TargetID   string         `json:"target_id"`
	Detail     map[string]any `json:"detail,omitempty"`
	CreatedAt  string         `json:"created_at"`
}

// ListAudit handles GET /admin/audit.
func (h *AdminHandler) ListAudit(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	rows, err := h.svc.ListAudit(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	dtos := make([]auditDTO, 0, len(rows))
	for _, e := range rows {
		dtos = append(dtos, auditDTO{
			ID: e.ID, Actor: e.Actor, Action: e.Action,
			TargetType: e.TargetType, TargetID: e.TargetID,
			Detail: e.Detail, CreatedAt: e.CreatedAt.Format(httpx.TimeFormat),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, dtos)
}
