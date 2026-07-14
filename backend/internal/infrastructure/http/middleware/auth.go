package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
)

// contextKey avoids collisions with other packages' context keys.
type contextKey string

// userIDContextKey is where RequireAuth stores the authenticated user's ID.
const userIDContextKey contextKey = "user_id"

// TokenVerifier is the minimal capability RequireAuth needs — just access
// token parsing — so the middleware depends on an interface, not the
// concrete infrastructure/security type (Dependency Injection).
type TokenVerifier interface {
	ParseAccessToken(token string) (user.ID, error)
}

// RequireAuth extracts and verifies a Bearer access token, rejecting the
// request with 401 if absent or invalid, and otherwise injects the
// authenticated user ID into the request context.
func RequireAuth(tokens TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				writeAuthError(w, apperrors.New(apperrors.CodeUnauthorized, "missing bearer token"))
				return
			}

			userID, err := tokens.ParseAccessToken(token)
			if err != nil {
				writeAuthError(w, err)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}
	return token, true
}

// UserIDFromContext retrieves the authenticated user ID a handler runs on
// behalf of. Only valid downstream of RequireAuth.
func UserIDFromContext(ctx context.Context) (user.ID, bool) {
	id, ok := ctx.Value(userIDContextKey).(user.ID)
	return id, ok
}

func writeAuthError(w http.ResponseWriter, err error) {
	// Minimal local JSON error write (rather than importing the http
	// package's writeError) to avoid a middleware<->handler import cycle:
	// the http package imports middleware for RequireAuth, so middleware
	// cannot import http back.
	body, marshalErr := json.Marshal(struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{Code: string(apperrors.CodeOf(err)), Message: err.Error()},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	if marshalErr == nil {
		_, _ = w.Write(body)
	}
}
