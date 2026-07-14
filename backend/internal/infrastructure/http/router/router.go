// Package router wires every backend module's HTTP handler onto routes,
// using the standard library's http.ServeMux (Go 1.22+ supports method and
// wildcard patterns natively, e.g. "POST /rules"), so no third-party
// router was vendored (KISS/YAGNI — see the project's other
// dependency-free infrastructure for the same rationale).
package router

import (
	"net/http"
	"time"

	"streampass/backend/internal/infrastructure/http/handler"
	"streampass/backend/internal/infrastructure/http/middleware"
	"streampass/shared/logger"
)

// Deps carries every handler and cross-cutting dependency the router needs
// to wire up. A single struct keeps main.go's construction call readable
// (Dependency Injection: composition root pattern).
type Deps struct {
	Auth      *handler.AuthHandler
	Rule      *handler.RuleHandler
	Relay     *handler.RelayHandler
	Telemetry *handler.TelemetryHandler
	Config    *handler.ConfigHandler
	Billing   *handler.BillingHandler
	Health    *handler.HealthHandler

	TokenVerifier middleware.TokenVerifier
	AdminKey      string
	Log           *logger.Logger

	// PublicRateLimit / AuthRateLimit let stricter limits apply to
	// unauthenticated, credential-guessing-prone endpoints (login,
	// register) than to normal authenticated traffic.
	PublicRateLimit int
	PublicWindow    time.Duration
}

// New builds the fully-wired root http.Handler.
func New(d Deps) http.Handler {
	mux := http.NewServeMux()

	authMW := middleware.RequireAuth(d.TokenVerifier)
	adminMW := middleware.RequireAdminKey(d.AdminKey)
	strictLimiter := middleware.NewRateLimiter(d.PublicRateLimit, d.PublicWindow)

	// --- Public, unauthenticated endpoints ---
	mux.HandleFunc("GET /health", d.Health.Check)
	mux.HandleFunc("GET /rules", d.Rule.GetLatest)
	mux.HandleFunc("GET /servers", d.Relay.ListAvailable)
	mux.HandleFunc("GET /config", d.Config.GetLatest)
	mux.Handle("POST /payments/webhook", strictLimiter.Middleware()(http.HandlerFunc(d.Billing.HandleWebhook)))

	// --- Auth endpoints (stricter rate limit: brute-force resistance) ---
	mux.Handle("POST /register", strictLimiter.Middleware()(http.HandlerFunc(d.Auth.Register)))
	mux.Handle("POST /login", strictLimiter.Middleware()(http.HandlerFunc(d.Auth.Login)))
	mux.Handle("POST /logout", strictLimiter.Middleware()(http.HandlerFunc(d.Auth.Logout)))

	// --- Authenticated client endpoints ---
	mux.Handle("POST /telemetry", authMW(http.HandlerFunc(d.Telemetry.Record)))
	mux.Handle("POST /payments", authMW(http.HandlerFunc(d.Billing.CreatePayment)))
	mux.Handle("GET /subscription", authMW(http.HandlerFunc(d.Billing.GetSubscription)))
	mux.Handle("POST /subscription/cancel", authMW(http.HandlerFunc(d.Billing.CancelSubscription)))

	// --- Admin / operator-only endpoints ---
	mux.Handle("POST /rules", adminMW(http.HandlerFunc(d.Rule.Publish)))
	mux.Handle("POST /config", adminMW(http.HandlerFunc(d.Config.Publish)))
	mux.Handle("POST /servers/health", adminMW(http.HandlerFunc(d.Relay.RecordHealthCheck)))

	return middleware.Chain(
		middleware.Recover(d.Log),
		middleware.Logging(d.Log),
	)(mux)
}
