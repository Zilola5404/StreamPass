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
	// UpdateHealth records a fresh health-check result for a relay,
	// written by the Health Monitor component.
	UpdateHealth(ctx context.Context, id ID, healthy bool, loadRatio float64, rttMillis int, checkedAt time.Time) error
}
