// Package diag is the operator diagnostic channel: per-flow routing
// outcomes (site/host, dest IP, DIRECT/RELAY, latency, decision reason, speed).
// Hostname / https://host origin only — never full URLs or page paths (ТЗ §14).
package diag

import (
	"context"
	"time"
)

// Event is one routing/connection diagnostic sample.
type Event struct {
	UserID          string
	Proto           string // tcp | udp | dns | vpn
	Site            string // https://host or ip://x.x.x.x (no path)
	Host            string
	DestIP          string
	DestPort        int
	Mode            string // DIRECT | RELAY | FALLBACK | DNS
	Result          string // ok | fail | timeout | drop | slow | xfer
	LatencyMS       int
	Slow            bool
	SpeedKbps       int
	Reason          string
	Rule            string
	DecisionReason  string
	ErrorCode       string
	RelayID         string
	ClientVersion   string
	RecordedAt      time.Time
}

// Repository persists and lists diagnostic events.
type Repository interface {
	RecordBatch(ctx context.Context, events []Event) error
	List(ctx context.Context, userID string, limit int) ([]Event, error)
	PurgeOlderThan(ctx context.Context, before time.Time) (int64, error)
}
