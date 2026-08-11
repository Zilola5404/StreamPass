// Package handler contains HTTP handlers for every backend module. Each
// handler only decodes the request, calls one application-layer use case,
// and encodes the response — no business logic (Clean Architecture:
// Interface/Delivery layer).
package handler

import (
	"net/http"

	"streampass/backend/internal/application/auth"
	httpx "streampass/backend/internal/infrastructure/http"
	"streampass/backend/internal/infrastructure/http/middleware"
)

// AuthHandler exposes /register, /login, /logout, /refresh, /me, password flows.
type AuthHandler struct {
	svc *auth.Service
}

// NewAuthHandler builds the Auth HTTP handler.
func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceID   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
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

	pair, err := h.svc.Login.Execute(r.Context(), auth.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	writeTokenPair(w, http.StatusCreated, pair)
}

type loginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceID   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
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

	pair, err := h.svc.Login.Execute(r.Context(), auth.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
	})
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

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken     string `json:"access_token"`
	AccessExpiresAt string `json:"access_expires_at"`
}

// Refresh handles "POST /refresh".
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	access, err := h.svc.Refresh.Execute(r.Context(), req.RefreshToken)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, refreshResponse{
		AccessToken:     access.AccessToken,
		AccessExpiresAt: access.AccessExpiresAt.Format(httpx.TimeFormat),
	})
}

type profileResponse struct {
	Email                   string  `json:"email"`
	CreatedAt               string  `json:"created_at"`
	SubscriptionActiveUntil *string `json:"subscription_active_until,omitempty"`
}

// GetProfile handles "GET /me" (authenticated).
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	profile, err := h.svc.GetProfile.Execute(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	resp := profileResponse{
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt.Format(httpx.TimeFormat),
	}
	if profile.SubscriptionActiveUntil != nil {
		formatted := profile.SubscriptionActiveUntil.Format(httpx.TimeFormat)
		resp.SubscriptionActiveUntil = &formatted
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword handles "PUT /me/password" (authenticated).
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	var req changePasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.svc.ChangePassword.Execute(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

// DeleteAccount handles "DELETE /me" (authenticated).
func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	if err := h.svc.DeleteAccount.Execute(r.Context(), userID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type forgotPasswordResponse struct {
	Message    string `json:"message"`
	ResetToken string `json:"reset_token,omitempty"`
}

// ForgotPassword handles "POST /password/forgot" (public, rate-limited).
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	result, err := h.svc.ForgotPassword.Execute(r.Context(), req.Email)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, forgotPasswordResponse{
		Message:    result.Message,
		ResetToken: result.ResetToken,
	})
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword handles "POST /password/reset" (public, rate-limited).
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.svc.ResetPassword.Execute(r.Context(), req.Token, req.NewPassword); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

type deviceResponse struct {
	ID         string `json:"id"`
	DeviceID   string `json:"device_id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
	LastSeenAt string `json:"last_seen_at"`
}

type devicesListResponse struct {
	Devices    []deviceResponse `json:"devices"`
	MaxDevices int              `json:"max_devices"`
}

// ListDevices handles "GET /me/devices" (authenticated).
func (h *AuthHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	list, err := h.svc.ListDevices.Execute(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	resp := devicesListResponse{
		Devices:    make([]deviceResponse, 0, len(list.Devices)),
		MaxDevices: list.MaxDevices,
	}
	for _, d := range list.Devices {
		resp.Devices = append(resp.Devices, deviceResponse{
			ID:         d.ID,
			DeviceID:   d.DeviceID,
			Name:       d.Name,
			CreatedAt:  d.CreatedAt.Format(httpx.TimeFormat),
			LastSeenAt: d.LastSeenAt.Format(httpx.TimeFormat),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// RevokeDevice handles "DELETE /me/devices/{id}" (authenticated).
func (h *AuthHandler) RevokeDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthenticated())
		return
	}
	id := r.PathValue("id")
	if id == "" {
		httpx.WriteError(w, httpx.ErrMissingPathValue("id"))
		return
	}
	if err := h.svc.RevokeDevice.Execute(r.Context(), userID, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}
