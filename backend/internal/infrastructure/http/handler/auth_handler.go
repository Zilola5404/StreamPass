// Package handler contains HTTP handlers for every backend module. Each
// handler only decodes the request, calls one application-layer use case,
// and encodes the response — no business logic (Clean Architecture:
// Interface/Delivery layer).
package handler

import (
	"net/http"

	"streampass/backend/internal/application/auth"
	httpx "streampass/backend/internal/infrastructure/http"
)

// AuthHandler exposes /register, /login, /logout.
type AuthHandler struct {
	svc *auth.Service
}

// NewAuthHandler builds the Auth HTTP handler.
func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles "POST /register".
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	if _, err := h.svc.Register.Execute(r.Context(), req.Email, req.Password); err != nil {
		httpx.WriteError(w, err)
		return
	}

	pair, err := h.svc.Login.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	writeTokenPair(w, http.StatusCreated, pair)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	AccessExpiresAt  string `json:"access_expires_at"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresAt string `json:"refresh_expires_at"`
}

// Login handles "POST /login".
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	pair, err := h.svc.Login.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	writeTokenPair(w, http.StatusOK, pair)
}

func writeTokenPair(w http.ResponseWriter, status int, pair *auth.TokenPair) {
	httpx.WriteJSON(w, status, tokenResponse{
		AccessToken:      pair.AccessToken,
		AccessExpiresAt:  pair.AccessExpiresAt.Format(httpx.TimeFormat),
		RefreshToken:     pair.RefreshToken,
		RefreshExpiresAt: pair.RefreshExpiresAt.Format(httpx.TimeFormat),
	})
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout handles "POST /logout".
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err := h.svc.Logout.Execute(r.Context(), req.RefreshToken); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
}
