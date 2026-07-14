// Package postgres holds infrastructure-layer implementations of
// repository ports backed by PostgreSQL, using database/sql + lib/pq.
package postgres

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq" // registers the "postgres" database/sql driver

	apperrors "streampass/shared/errors"
)

// PoolConfig configures the underlying *sql.DB connection pool.
type PoolConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Open establishes a PostgreSQL connection pool and verifies connectivity.
func Open(cfg PoolConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeUnavailable, "failed to open postgres connection", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeUnavailable, "failed to reach postgres", err)
	}

	return db, nil
}
