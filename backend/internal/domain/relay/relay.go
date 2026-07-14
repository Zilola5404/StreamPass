// Package relay contains the Relay Manager's domain model: the registry of
// relay servers exposed to clients via GET /servers (spec section 9) and
// updated by the Health Monitor (spec section 4).
package relay

import "time"

// ID uniquely identifies a relay server.
type ID string

// Region is a human-readable relay location, per the spec's starting
// infrastructure (section 9: Germany/Frankfurt, Netherlands/Amsterdam).
type Region string

// Server is a single Hysteria2 relay endpoint.
type Server struct {
	ID        ID
	Region    Region
	Host      string
	Port      int
	Healthy   bool
	LoadRatio float64 // 0.0 (idle) .. 1.0 (full capacity)
	RTTMillis int
	UpdatedAt time.Time
}

// IsSelectable reports whether a server should be offered to clients right
// now. Kept as a domain method (not scattered comparisons in the
// application layer) so the "what counts as usable" rule has one home.
func (s *Server) IsSelectable() bool {
	return s.Healthy
}
