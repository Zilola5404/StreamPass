package postgres

import (
	"context"
	"database/sql"
	"time"

	"streampass/backend/internal/domain/diag"
	apperrors "streampass/shared/errors"
)

// DiagRepository implements diag.Repository.
type DiagRepository struct {
	db *sql.DB
}

// NewDiagRepository builds a Postgres-backed diag.Repository.
func NewDiagRepository(db *sql.DB) *DiagRepository {
	return &DiagRepository{db: db}
}

// RecordBatch inserts many diagnostic events in one transaction.
func (r *DiagRepository) RecordBatch(ctx context.Context, events []diag.Event) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin diag tx", err)
	}
	defer func() { _ = tx.Rollback() }()

	const q = `
		INSERT INTO diag_events
			(user_id, proto, site, host, dest_ip, dest_port, mode, result, latency_ms, slow, speed_kbps, reason, rule, decision_reason, error_code, relay_id, client_version, recorded_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`

	for _, e := range events {
		if _, err := tx.ExecContext(ctx, q,
			e.UserID, e.Proto, e.Site, e.Host, e.DestIP, e.DestPort, e.Mode, e.Result,
			e.LatencyMS, e.Slow, e.SpeedKbps, e.Reason, e.Rule, e.DecisionReason,
			e.ErrorCode, e.RelayID, e.ClientVersion, e.RecordedAt,
		); err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "failed to insert diag event", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to commit diag tx", err)
	}
	return nil
}

// List returns newest diagnostic events, optionally filtered by user.
func (r *DiagRepository) List(ctx context.Context, userID string, limit int) ([]diag.Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var (
		rows *sql.Rows
		err  error
	)
	const cols = `user_id, proto, site, host, dest_ip, dest_port, mode, result, latency_ms, slow, speed_kbps, reason, rule, decision_reason, error_code, relay_id, client_version, recorded_at`
	if userID != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT `+cols+` FROM diag_events WHERE user_id = $1 ORDER BY recorded_at DESC LIMIT $2`,
			userID, limit)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT `+cols+` FROM diag_events ORDER BY recorded_at DESC LIMIT $1`,
			limit)
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list diag events", err)
	}
	defer rows.Close()

	var out []diag.Event
	for rows.Next() {
		var e diag.Event
		if err := rows.Scan(
			&e.UserID, &e.Proto, &e.Site, &e.Host, &e.DestIP, &e.DestPort, &e.Mode, &e.Result,
			&e.LatencyMS, &e.Slow, &e.SpeedKbps, &e.Reason, &e.Rule, &e.DecisionReason,
			&e.ErrorCode, &e.RelayID, &e.ClientVersion, &e.RecordedAt,
		); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan diag event", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed while iterating diag events", err)
	}
	if out == nil {
		out = []diag.Event{}
	}
	return out, nil
}

// PurgeOlderThan deletes events older than before.
func (r *DiagRepository) PurgeOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM diag_events WHERE recorded_at < $1`, before)
	if err != nil {
		return 0, apperrors.Wrap(apperrors.CodeInternal, "failed to purge diag events", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
