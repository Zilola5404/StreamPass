package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
)

// ExclusionRepository stores per-user DIRECT domain exclusions as JSONB.
type ExclusionRepository struct {
	db *sql.DB
}

// NewExclusionRepository builds a Postgres-backed exclusion repository.
func NewExclusionRepository(db *sql.DB) *ExclusionRepository {
	return &ExclusionRepository{db: db}
}

// Get returns the user's exclusions (empty slice when unset).
func (r *ExclusionRepository) Get(ctx context.Context, userID user.ID) ([]string, error) {
	const q = `SELECT exclusions FROM users WHERE id = $1`
	var raw []byte
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, apperrors.New(apperrors.CodeNotFound, "user not found")
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to read exclusions", err)
	}
	if len(raw) == 0 {
		return []string{}, nil
	}
	var domains []string
	if err := json.Unmarshal(raw, &domains); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to decode exclusions", err)
	}
	if domains == nil {
		return []string{}, nil
	}
	return domains, nil
}

// Replace overwrites the user's exclusions list.
func (r *ExclusionRepository) Replace(ctx context.Context, userID user.ID, domains []string) error {
	if domains == nil {
		domains = []string{}
	}
	raw, err := json.Marshal(domains)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to encode exclusions", err)
	}
	const q = `
		UPDATE users
		SET exclusions = $2::jsonb, updated_at = NOW()
		WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, userID, raw)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update exclusions", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to read rows affected", err)
	}
	if n == 0 {
		return apperrors.New(apperrors.CodeNotFound, "user not found")
	}
	return nil
}
