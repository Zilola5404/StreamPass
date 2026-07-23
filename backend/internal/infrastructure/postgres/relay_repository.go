package postgres

import (
	"context"
	"database/sql"
	"time"

	"streampass/backend/internal/domain/relay"
	apperrors "streampass/shared/errors"
)

type RelayRepository struct {
	db *sql.DB
}

func NewRelayRepository(db *sql.DB) *RelayRepository {
	return &RelayRepository{db: db}
}

func (r *RelayRepository) List(ctx context.Context) ([]relay.Server, error) {
	const q = `
		SELECT id, region, host, port, healthy, load_ratio, rtt_millis, connection_config, updated_at
		FROM relay_servers`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list relay servers", err)
	}
	defer rows.Close()
	var servers []relay.Server
	for rows.Next() {
		var s relay.Server
		if err := rows.Scan(&s.ID, &s.Region, &s.Host, &s.Port, &s.Healthy, &s.LoadRatio, &s.RTTMillis, &s.ConnectionConfig, &s.UpdatedAt); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan relay server row", err)
		}
		servers = append(servers, s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed while iterating relay servers", err)
	}
	return servers, nil
}

func (r *RelayRepository) Register(ctx context.Context, s relay.Server) (*relay.Server, error) {
	const q = `
		INSERT INTO relay_servers (id, region, host, port, healthy, load_ratio, rtt_millis, connection_config, updated_at)
		VALUES ($1, $2, $3, $4, false, 0, 0, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			region = EXCLUDED.region,
			host = EXCLUDED.host,
			port = EXCLUDED.port,
			connection_config = EXCLUDED.connection_config
		RETURNING id, region, host, port, healthy, load_ratio, rtt_millis, connection_config, updated_at`

	var result relay.Server
	err := r.db.QueryRowContext(ctx, q, s.ID, s.Region, s.Host, s.Port, s.ConnectionConfig, s.UpdatedAt).Scan(
		&result.ID, &result.Region, &result.Host, &result.Port,
		&result.Healthy, &result.LoadRatio, &result.RTTMillis, &result.ConnectionConfig, &result.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to register relay server", err)
	}
	return &result, nil
}

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

func (r *RelayRepository) Delete(ctx context.Context, id relay.ID) error {
	const q = `DELETE FROM relay_servers WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to delete relay server", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm relay server deletion", err)
	}
	if n == 0 {
		return apperrors.New(apperrors.CodeNotFound, "relay server not found").WithDetails(map[string]any{"id": id})
	}
	return nil
}
