package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
)

// pqUniqueViolationCode is the Postgres SQLSTATE for a unique-constraint
// violation, used to translate a duplicate email insert into a
// domain-meaningful AppError.
const pqUniqueViolationCode = "23505"

const userSelectCols = `id, email, password_hash, created_at, updated_at, subscription_active_until, banned_at`

// UserRepository implements user.Repository against the "users" table.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository builds a Postgres-backed user.Repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user row.
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	const q = `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, q, u.ID, u.Email, u.PasswordHash, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.New(apperrors.CodeAlreadyExists, "email already registered")
		}
		return apperrors.Wrap(apperrors.CodeInternal, "failed to insert user", err)
	}
	return nil
}

// FindByEmail looks up a user by email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	q := `SELECT ` + userSelectCols + ` FROM users WHERE email = $1`
	return r.scanOne(r.db.QueryRowContext(ctx, q, email), email)
}

// FindByID looks up a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	q := `SELECT ` + userSelectCols + ` FROM users WHERE id = $1`
	return r.scanOne(r.db.QueryRowContext(ctx, q, id), string(id))
}

func (r *UserRepository) scanOne(row *sql.Row, lookupKey string) (*user.User, error) {
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrNotFound(lookupKey)
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan user row", err)
	}
	return u, nil
}

func scanUser(scanner interface {
	Scan(dest ...any) error
}) (*user.User, error) {
	var u user.User
	var subUntil, bannedAt sql.NullTime
	err := scanner.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt, &subUntil, &bannedAt)
	if err != nil {
		return nil, err
	}
	if subUntil.Valid {
		u.SubscriptionActiveUntil = &subUntil.Time
	}
	if bannedAt.Valid {
		u.BannedAt = &bannedAt.Time
	}
	return &u, nil
}

// isUniqueViolation detects Postgres unique-constraint violations via the
// driver's typed error rather than string matching.
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == pqUniqueViolationCode
}

// ExtendSubscription sets the user's subscription expiry timestamp.
func (r *UserRepository) ExtendSubscription(ctx context.Context, id user.ID, activeUntil time.Time) error {
	const q = `UPDATE users SET subscription_active_until = $2, updated_at = NOW() WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id, activeUntil)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to extend subscription", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm subscription update", err)
	}
	if n == 0 {
		return user.ErrNotFound(string(id))
	}
	return nil
}

// ClearSubscription removes Premium access.
func (r *UserRepository) ClearSubscription(ctx context.Context, id user.ID, now time.Time) error {
	const q = `UPDATE users SET subscription_active_until = NULL, updated_at = $2 WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to clear subscription", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm subscription clear", err)
	}
	if n == 0 {
		return user.ErrNotFound(string(id))
	}
	return nil
}

// SetBanned sets or clears banned_at (nil = unban).
func (r *UserRepository) SetBanned(ctx context.Context, id user.ID, bannedAt *time.Time, now time.Time) error {
	const q = `UPDATE users SET banned_at = $2, updated_at = $3 WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id, bannedAt, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update ban status", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm ban update", err)
	}
	if n == 0 {
		return user.ErrNotFound(string(id))
	}
	return nil
}

// UpdatePasswordHash replaces the password hash and bumps updated_at.
func (r *UserRepository) UpdatePasswordHash(ctx context.Context, id user.ID, passwordHash string, now time.Time) error {
	const q = `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id, passwordHash, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update password", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm password update", err)
	}
	if n == 0 {
		return user.ErrNotFound(string(id))
	}
	return nil
}

// Delete removes the user and cascading payment rows in one transaction.
func (r *UserRepository) Delete(ctx context.Context, id user.ID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin delete user tx", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM payments WHERE user_id = $1`, id); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to delete user payments", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_devices WHERE user_id = $1`, id); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to delete user devices", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to delete user", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm user delete", err)
	}
	if n == 0 {
		return user.ErrNotFound(string(id))
	}
	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to commit delete user", err)
	}
	return nil
}

// List returns every registered user, newest first.
func (r *UserRepository) List(ctx context.Context) ([]*user.User, error) {
	return r.SearchByEmail(ctx, "")
}

// SearchByEmail returns users matching email substring (ILIKE), or all when q empty.
func (r *UserRepository) SearchByEmail(ctx context.Context, q string) ([]*user.User, error) {
	q = strings.TrimSpace(q)
	var rows *sql.Rows
	var err error
	if q == "" {
		rows, err = r.db.QueryContext(ctx, `SELECT `+userSelectCols+` FROM users ORDER BY created_at DESC`)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT `+userSelectCols+` FROM users WHERE email ILIKE '%' || $1 || '%' ORDER BY created_at DESC`, q)
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list users", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan user row", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed while iterating users", err)
	}
	if users == nil {
		users = []*user.User{}
	}
	return users, nil
}
