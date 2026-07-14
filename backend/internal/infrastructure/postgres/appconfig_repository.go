package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"streampass/backend/internal/domain/appconfig"
	apperrors "streampass/shared/errors"
)

// AppConfigRepository implements appconfig.Repository against the
// "app_configs" table, one immutable row per published version — the same
// versioning pattern as RuleRepository (spec: Config Service mirrors Rule
// Service's publish/version model).
type AppConfigRepository struct {
	db *sql.DB
}

// NewAppConfigRepository builds a Postgres-backed appconfig.Repository.
func NewAppConfigRepository(db *sql.DB) *AppConfigRepository {
	return &AppConfigRepository{db: db}
}

// Latest returns the highest-versioned config.
func (r *AppConfigRepository) Latest(ctx context.Context) (*appconfig.Config, error) {
	const q = `
		SELECT version, min_supported_client_version, telemetry_enabled,
		       rule_poll_interval_sec, relay_poll_interval_sec, updated_at
		FROM app_configs
		ORDER BY version DESC
		LIMIT 1`

	var c appconfig.Config
	err := r.db.QueryRowContext(ctx, q).Scan(
		&c.Version, &c.MinSupportedClientVer, &c.TelemetryEnabled,
		&c.RulePollIntervalSec, &c.RelayPollIntervalSec, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, appconfig.ErrNoConfig
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to load latest config", err)
	}
	return &c, nil
}

// Publish inserts a new version (previous max + 1) inside a transaction to
// avoid a race between two concurrent publishers picking the same version.
func (r *AppConfigRepository) Publish(ctx context.Context, c appconfig.Config, publishedAt time.Time) (*appconfig.Config, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback()

	var nextVersion int
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) + 1 FROM app_configs`).Scan(&nextVersion)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to compute next config version", err)
	}

	const insert = `
		INSERT INTO app_configs
			(version, min_supported_client_version, telemetry_enabled,
			 rule_poll_interval_sec, relay_poll_interval_sec, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.ExecContext(ctx, insert,
		nextVersion, c.MinSupportedClientVer, c.TelemetryEnabled,
		c.RulePollIntervalSec, c.RelayPollIntervalSec, publishedAt,
	)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to insert config", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to commit config publish", err)
	}

	c.Version = nextVersion
	c.UpdatedAt = publishedAt
	return &c, nil
}
