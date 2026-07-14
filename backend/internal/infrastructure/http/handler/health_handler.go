package handler

import (
	"net/http"

	httpx "streampass/backend/internal/infrastructure/http"
)

// HealthHandler exposes GET /health for load balancer / orchestrator
// liveness probes. Deliberately has no dependencies on the database or
// Redis: it answers "is this process alive", not "are its dependencies
// healthy" — a DB outage should not make the load balancer kill every
// backend instance simultaneously.
type HealthHandler struct{}

// NewHealthHandler builds the health check handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check handles "GET /health".
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
