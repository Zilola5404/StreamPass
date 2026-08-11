package user

import (
	"context"
	"time"
)

// AuditEntry records a sensitive admin action (BL-050).
type AuditEntry struct {
	ID         string
	Actor      string
	Action     string
	TargetType string
	TargetID   string
	Detail     map[string]any
	CreatedAt  time.Time
}

// AuditRepository persists admin audit events.
type AuditRepository interface {
	Append(ctx context.Context, e *AuditEntry) error
	List(ctx context.Context, limit int) ([]*AuditEntry, error)
}
