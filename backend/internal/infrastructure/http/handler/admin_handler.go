package handler

import (
	"net/http"

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
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	dtos := make([]userSummaryDTO, len(users))
	for i, u := range users {
		dto := userSummaryDTO{
			ID:                 string(u.ID),
			Email:              u.Email,
			CreatedAt:          u.CreatedAt.Format(httpx.TimeFormat),
			SubscriptionActive: u.IsSubscriptionActive,
		}
		if u.SubscriptionActiveUntil != nil {
			formatted := u.SubscriptionActiveUntil.Format(httpx.TimeFormat)
			dto.SubscriptionActiveUntil = &formatted
		}
		dtos[i] = dto
	}
	httpx.WriteJSON(w, http.StatusOK, dtos)
}
