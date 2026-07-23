package relay

import (
	"context"
	"time"
)

// Repository is the port the Relay Manager application layer depends on.
type Repository interface {
	// List returns every known relay server (healthy or not); filtering
	// for client responses is an application-layer concern.
	List(ctx context.Context) ([]Server, error)
	// Register creates a new relay server entry, or updates its connection
	// details (region/host/port) if the ID already exists. Health status is
	// left untouched on update — registration is a separate concern from
	// health reporting (see UpdateHealth).
	Register(ctx context.Context, s Server) (*Server, error)
	// UpdateHealth records a fresh health-check result for a relay,
	// written by the Health Monitor component.
	UpdateHealth(ctx context.Context, id ID, healthy bool, loadRatio float64, rttMillis int, checkedAt time.Time) error
	// Delete removes a relay server from the registry (e.g. decommissioning,
	// or cleaning up a mistaken registration).
	Delete(ctx context.Context, id ID) error
}
