package postgres

import (
	"context"
	"database/sql"

	"streampass/backend/internal/domain/telemetry"
	apperrors "streampass/shared/errors"
)

// TelemetryRepository implements telemetry.Repository against the
// "telemetry_events" table. Write-only from the application's
// perspective — nothing in the backend reads this table back out to
// clients; it feeds Prometheus/Grafana dashboards (spec section 18) via a
// separate exporter, not modeled in this MVP backend.
type TelemetryRepository struct {
	db *sql.DB
}

// NewTelemetryRepository builds a Postgres-backed telemetry.Repository.
func NewTelemetryRepository(db *sql.DB) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}

// Record inserts one telemetry event row.
func (r *TelemetryRepository) Record(ctx context.Context, e telemetry.Event) error {
	const q = `
		INSERT INTO telemetry_events
			(user_id, rtt_millis, packet_loss_pct, relay_id, client_version, os, connect_millis, error_code, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, q,
		e.UserID, e.RTTMillis, e.PacketLossPct, e.RelayID, e.ClientVersion, e.OS, e.ConnectMillis, e.ErrorCode, e.RecordedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to insert telemetry event", err)
	}
	return nil
}
