package postgres

import (
	"context"
	"database/sql"
	"time"

	"streampass/backend/internal/domain/relay"
	apperrors "streampass/shared/errors"
)

// RelayRepository implements relay.Repository against the "relay_servers"
// table.
type RelayRepository struct {
	db *sql.DB
}

// NewRelayRepository builds a Postgres-backed relay.Repository.
func NewRelayRepository(db *sql.DB) *RelayRepository {
	return &RelayRepository{db: db}
}

// List returns every known relay server.
func (r *RelayRepository) List(ctx context.Context) ([]relay.Server, error) {
	const q = `
		SELECT id, region, host, port, healthy, load_ratio, rtt_millis, updated_at
		FROM relay_servers`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list relay servers", err)
	}
	defer rows.Close()

	var servers []relay.Server
	for rows.Next() {
		var s relay.Server
		if err := rows.Scan(&s.ID, &s.Region, &s.Host, &s.Port, &s.Healthy, &s.LoadRatio, &s.RTTMillis, &s.UpdatedAt); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan relay server row", err)
		}
		servers = append(servers, s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed while iterating relay servers", err)
	}
	return servers, nil
}

// UpdateHealth records a fresh health-check result for a relay.
func (r *RelayRepository) UpdateHealth(ctx context.Context, id relay.ID, healthy bool, loadRatio float64, rttMillis int, checkedAt time.Time) error {
	const q = `
		UPDATE relay_servers
		SET healthy = $2, load_ratio = $3, rtt_millis = $4, updated_at = $5
		WHERE id = $1`

	res, err := r.db.ExecContext(ctx, q, id, healthy, loadRatio, rttMillis, checkedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update relay health", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm relay health update", err)
	}
	if n == 0 {
		return apperrors.New(apperrors.CodeNotFound, "relay server not found").WithDetails(map[string]any{"id": id})
	}
	return nil
}
