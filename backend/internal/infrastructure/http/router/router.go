// Package router wires every backend module's HTTP handler onto routes,
// using the standard library's http.ServeMux (Go 1.22+ supports method and
// wildcard patterns natively, e.g. "POST /rules"), so no third-party
// router was vendored (KISS/YAGNI — see the project's other
// dependency-free infrastructure for the same rationale).
package router

import (
	"net/http"
	"strings"
	"time"

	"streampass/backend/internal/infrastructure/http/handler"
	"streampass/backend/internal/infrastructure/http/middleware"
	"streampass/shared/logger"
)

// apiV1Prefix versions every business endpoint per ТЗ §13 ("Все API
// должны иметь версионирование (/api/v1/...)").
const apiV1Prefix = "/api/v1"

// v1 rewrites a ServeMux pattern like "GET /rules" into
// "GET /api/v1/rules" — every route below states its path once, and this
// keeps the prefix change in exactly one place instead of hand-editing
// every literal string.
func v1(pattern string) string {
	method, path, ok := strings.Cut(pattern, " ")
	if !ok {
		return apiV1Prefix + pattern
	}
	return method + " " + apiV1Prefix + path
}

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
	Admin     *handler.AdminHandler

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
	// /health is intentionally available both bare (for load balancer /
	// orchestrator probes, which conventionally expect an unversioned
	// path) and under /api/v1 (for spec-literal compliance) — the two
	// share one handler, so there's nothing to keep in sync.
	mux.HandleFunc("GET /health", d.Health.Check)
	mux.HandleFunc(v1("GET /health"), d.Health.Check)
	mux.HandleFunc(v1("GET /rules"), d.Rule.GetLatest)
	mux.HandleFunc(v1("GET /config"), d.Config.GetLatest)
	mux.Handle(v1("POST /payments/webhook"), strictLimiter.Middleware()(http.HandlerFunc(d.Billing.HandleWebhook)))

	// --- Auth endpoints (stricter rate limit: brute-force resistance) ---
	mux.Handle(v1("POST /register"), strictLimiter.Middleware()(http.HandlerFunc(d.Auth.Register)))
	mux.Handle(v1("POST /login"), strictLimiter.Middleware()(http.HandlerFunc(d.Auth.Login)))
	mux.Handle(v1("POST /logout"), strictLimiter.Middleware()(http.HandlerFunc(d.Auth.Logout)))
	mux.Handle(v1("POST /refresh"), strictLimiter.Middleware()(http.HandlerFunc(d.Auth.Refresh)))

	// --- Authenticated client endpoints ---
	// GET /servers now returns each relay's ConnectionConfig (a real VPN
	// connection secret), so — unlike the other GET endpoints above — it
	// must never be reachable without a valid access token.
	mux.Handle(v1("GET /servers"), authMW(http.HandlerFunc(d.Relay.ListAvailable)))
	mux.Handle(v1("POST /telemetry"), authMW(http.HandlerFunc(d.Telemetry.Record)))
	mux.Handle(v1("POST /payments"), authMW(http.HandlerFunc(d.Billing.CreatePayment)))
	mux.Handle(v1("GET /subscription"), authMW(http.HandlerFunc(d.Billing.GetSubscription)))
	mux.Handle(v1("POST /subscription/cancel"), authMW(http.HandlerFunc(d.Billing.CancelSubscription)))

	// --- Admin / operator-only endpoints ---
	mux.Handle(v1("POST /rules"), adminMW(http.HandlerFunc(d.Rule.Publish)))
	mux.Handle(v1("POST /config"), adminMW(http.HandlerFunc(d.Config.Publish)))
	mux.Handle(v1("GET /servers/all"), adminMW(http.HandlerFunc(d.Relay.ListAll)))
	mux.Handle(v1("POST /servers"), adminMW(http.HandlerFunc(d.Relay.Register)))
	mux.Handle(v1("DELETE /servers/{id}"), adminMW(http.HandlerFunc(d.Relay.Delete)))
	mux.Handle(v1("POST /servers/health"), adminMW(http.HandlerFunc(d.Relay.RecordHealthCheck)))
	mux.Handle(v1("GET /users"), adminMW(http.HandlerFunc(d.Admin.ListUsers)))

	return middleware.Chain(
		middleware.Recover(d.Log),
		middleware.Logging(d.Log),
	)(mux)
}
