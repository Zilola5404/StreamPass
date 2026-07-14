// Package telemetry contains the domain model for technical telemetry
// events (spec section 14). Only technical connection parameters are
// modeled here — no browsing history, no traffic contents, no URLs, no
// personal data — matching the spec's explicit allow/deny list.
package telemetry

import (
	"context"
	"time"
)

// Event is one technical telemetry sample from a client.
type Event struct {
	UserID        string
	RTTMillis     int
	PacketLossPct float64
	RelayID       string
	ClientVersion string
	OS            string
	ConnectMillis int
	ErrorCode     string // empty when the connection succeeded
	RecordedAt    time.Time
}

// Repository is the port the Telemetry application layer depends on.
type Repository interface {
	Record(ctx context.Context, e Event) error
}
