package relay

import "time"

// ID uniquely identifies a relay server.
type ID string

// Region is a human-readable relay location.
type Region string

// Server is a single relay endpoint (currently backed by Hiddify Manager:
// Xray/Reality, Hysteria2).
type Server struct {
	ID        ID
	Region    Region
	Host      string
	Port      int
	Healthy   bool
	LoadRatio float64
	RTTMillis int
	ConnectionConfig string
	UpdatedAt        time.Time
}

// IsSelectable reports whether a server should be offered to clients right
// now.
func (s *Server) IsSelectable() bool {
	return s.Healthy
}
