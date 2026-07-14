package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"streampass/backend/internal/domain/rule"
	apperrors "streampass/shared/errors"
)

// RuleRepository implements rule.Repository against the "rule_sets" table.
// Each row is one immutable published version, storing its rules as a
// JSONB array — simpler than a normalized rules table for MVP scale
// (a few hundred rules) and keeps the whole set atomic per version (KISS).
type RuleRepository struct {
	db *sql.DB
}

// NewRuleRepository builds a Postgres-backed rule.Repository.
func NewRuleRepository(db *sql.DB) *RuleRepository {
	return &RuleRepository{db: db}
}

type ruleRow struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
	Mode    string `json:"mode"`
}

// Latest returns the highest-versioned rule set.
func (r *RuleRepository) Latest(ctx context.Context) (*rule.Set, error) {
	const q = `
		SELECT version, rules, created_at
		FROM rule_sets
		ORDER BY version DESC
		LIMIT 1`

	var version int
	var rulesJSON []byte
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, q).Scan(&version, &rulesJSON, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, rule.ErrNoRuleSet
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to load latest rule set", err)
	}

	rules, err := decodeRules(rulesJSON)
	if err != nil {
		return nil, err
	}
	return rule.NewSet(version, rules, createdAt), nil
}

// Publish inserts a new version (previous max + 1) inside a transaction to
// avoid a race between two concurrent publishers picking the same version
// number.
func (r *RuleRepository) Publish(ctx context.Context, rules []rule.Rule, createdAt time.Time) (*rule.Set, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback()

	var nextVersion int
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) + 1 FROM rule_sets`).Scan(&nextVersion)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to compute next rule set version", err)
	}

	rulesJSON, err := encodeRules(rules)
	if err != nil {
		return nil, err
	}

	const insert = `INSERT INTO rule_sets (version, rules, created_at) VALUES ($1, $2, $3)`
	if _, err := tx.ExecContext(ctx, insert, nextVersion, rulesJSON, createdAt); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to insert rule set", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to commit rule set publish", err)
	}

	return rule.NewSet(nextVersion, rules, createdAt), nil
}

func encodeRules(rules []rule.Rule) ([]byte, error) {
	rows := make([]ruleRow, len(rules))
	for i, r := range rules {
		rows[i] = ruleRow{Kind: string(r.Kind), Pattern: r.Pattern, Mode: string(r.Mode)}
	}
	b, err := json.Marshal(rows)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to encode rules", err)
	}
	return b, nil
}

func decodeRules(data []byte) ([]rule.Rule, error) {
	var rows []ruleRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to decode rules", err)
	}
	rules := make([]rule.Rule, len(rows))
	for i, row := range rows {
		rules[i] = rule.Rule{Kind: rule.Kind(row.Kind), Pattern: row.Pattern, Mode: rule.Mode(row.Mode)}
	}
	return rules, nil
}
