package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"streampass/backend/internal/domain/user"
	"streampass/shared/idgen"
	apperrors "streampass/shared/errors"
)

// AuditRepository stores admin audit events (BL-050).
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository builds a Postgres-backed audit repository.
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Append inserts one audit row.
func (r *AuditRepository) Append(ctx context.Context, e *user.AuditEntry) error {
	if e.ID == "" {
		e.ID = idgen.New()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	if e.Detail == nil {
		e.Detail = map[string]any{}
	}
	raw, err := json.Marshal(e.Detail)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to encode audit detail", err)
	}
	const q = `
		INSERT INTO admin_audit_log (id, actor, action, target_type, target_id, detail, created_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)`
	_, err = r.db.ExecContext(ctx, q, e.ID, e.Actor, e.Action, e.TargetType, e.TargetID, raw, e.CreatedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to append audit log", err)
	}
	return nil
}

// List returns recent audit events, newest first.
func (r *AuditRepository) List(ctx context.Context, limit int) ([]*user.AuditEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	const q = `
		SELECT id, actor, action, target_type, target_id, detail, created_at
		FROM admin_audit_log
		ORDER BY created_at DESC
		LIMIT $1`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list audit log", err)
	}
	defer rows.Close()

	var out []*user.AuditEntry
	for rows.Next() {
		var e user.AuditEntry
		var raw []byte
		if err := rows.Scan(&e.ID, &e.Actor, &e.Action, &e.TargetType, &e.TargetID, &raw, &e.CreatedAt); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan audit row", err)
		}
		e.Detail = map[string]any{}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &e.Detail)
		}
		out = append(out, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to iterate audit log", err)
	}
	if out == nil {
		out = []*user.AuditEntry{}
	}
	return out, nil
}