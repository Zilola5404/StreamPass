package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	apperrors "streampass/shared/errors"
)

// RequireAdminKey protects operator-only endpoints (publishing rule sets,
// publishing config) with a static, out-of-band shared secret rather than
// a full user session — the spec's Admin Panel is a separate module; this
// is the minimal server-side gate its API calls must pass until that
// module exists. The key is read from config/environment, never hardcoded
// (spec: "Запрещено использовать hardcode").
func RequireAdminKey(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("X-Admin-Key")
			if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expectedKey)) != 1 {
				writeForbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeForbidden(w http.ResponseWriter) {
	body, err := json.Marshal(struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{Code: string(apperrors.CodeForbidden), Message: "invalid or missing admin key"},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	if err == nil {
		_, _ = w.Write(body)
	}
}
