// Package diag is the operator diagnostic channel: per-flow routing
// outcomes (host, dest IP, DIRECT/RELAY, latency, error). Stores hostname
// only — never full URLs or page paths (ТЗ §14 allow/deny).
package diag

import (
	"context"
	"time"
)

// Event is one routing/connection diagnostic sample.
type Event struct {
	UserID        string
	Proto         string // tcp | udp | dns | vpn
	Host          string // FQDN without scheme/path; may be empty for IP-only
	DestIP        string
	DestPort      int
	Mode          string // DIRECT | RELAY | FALLBACK | BLOCK | DROP
	Result        string // ok | fail | timeout | reject | drop
	LatencyMS     int
	ErrorCode     string
	RelayID       string
	ClientVersion string
	RecordedAt    time.Time
}

// Repository persists and lists diagnostic events.
type Repository interface {
	RecordBatch(ctx context.Context, events []Event) error
	List(ctx context.Context, userID string, limit int) ([]Event, error)
	PurgeOlderThan(ctx context.Context, before time.Time) (int64, error)
}
