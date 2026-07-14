// Package middleware contains HTTP middleware shared across all backend
// routes: request logging, panic recovery, JWT authentication and rate
// limiting.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"streampass/shared/logger"
)

// statusRecorder captures the status code written by the handler so the
// logging middleware can report it (http.ResponseWriter doesn't expose it).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Logging logs one structured line per request: method, path, status,
// duration. Applied outermost in the middleware chain so it captures the
// full request lifetime, including any panic recovered by Recover.
func Logging(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			log.Info(r.Context(), "http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// Recover catches panics in downstream handlers, logs them and returns a
// 500 instead of crashing the process — a single client-triggered bug must
// not take the whole server down (spec: reliability).
func Recover(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error(r.Context(), panicError{rec}, slog.String("path", r.URL.Path))
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":{"code":"INTERNAL","message":"internal error"}}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type panicError struct{ v any }

func (p panicError) Error() string { return "panic recovered: " + toString(p.v) }

func toString(v any) string {
	if err, ok := v.(error); ok {
		return err.Error()
	}
	if s, ok := v.(string); ok {
		return s
	}
	return "non-error panic value"
}

// Chain composes middleware in the order given: Chain(a, b, c)(h) runs as
// a(b(c(h))).
func Chain(mws ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		h := final
		for i := len(mws) - 1; i >= 0; i-- {
			h = mws[i](h)
		}
		return h
	}
}
