package postgres

import (
	"context"
	"database/sql"
	"embed"
	"sort"
	"strings"

	apperrors "streampass/shared/errors"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

const migrationsDir = "migrations"

// Migrate applies every *.up.sql migration that hasn't been recorded yet,
// in filename order, inside a schema_migrations tracking table. A minimal
// hand-rolled runner was used instead of vendoring golang-migrate: the
// project only ever applies forward migrations in one direction at
// startup, which this covers in ~40 lines (KISS/YAGNIP).
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create schema_migrations table", err)
	}

	entries, err := migrationFiles.ReadDir(migrationsDir)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to read embedded migrations", err)
	}

	var filenames []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			filenames = append(filenames, e.Name())
		}
	}
	sort.Strings(filenames)

	for _, name := range filenames {
		applied, err := isApplied(ctx, db, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := migrationFiles.ReadFile(migrationsDir + "/" + name)
		if err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "failed to read migration file: "+name, err)
		}

		if err := applyMigration(ctx, db, name, string(sqlBytes)); err != nil {
			return err
		}
	}
	return nil
}

func isApplied(ctx context.Context, db *sql.DB, filename string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`, filename).Scan(&exists)
	if err != nil {
		return false, apperrors.Wrap(apperrors.CodeInternal, "failed to check migration status: "+filename, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, db *sql.DB, filename, sqlText string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin migration transaction: "+filename, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, sqlText); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to apply migration: "+filename, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (filename) VALUES ($1)`, filename); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to record migration: "+filename, err)
	}
	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to commit migration: "+filename, err)
	}
	return nil
}
