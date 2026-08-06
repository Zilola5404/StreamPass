package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
)

// pqUniqueViolationCode is the Postgres SQLSTATE for a unique-constraint
// violation, used to translate a duplicate email insert into a
// domain-meaningful AppError.
const pqUniqueViolationCode = "23505"

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
	const q = `
		SELECT id, email, password_hash, created_at, updated_at, subscription_active_until
		FROM users WHERE email = $1`

	return r.scanOne(r.db.QueryRowContext(ctx, q, email), email)
}

// FindByID looks up a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at, updated_at, subscription_active_until
		FROM users WHERE id = $1`

	return r.scanOne(r.db.QueryRowContext(ctx, q, id), string(id))
}

func (r *UserRepository) scanOne(row *sql.Row, lookupKey string) (*user.User, error) {
	var u user.User
	var subUntil sql.NullTime

	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt, &subUntil)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrNotFound(lookupKey)
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan user row", err)
	}
	if subUntil.Valid {
		u.SubscriptionActiveUntil = &subUntil.Time
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
	const q = `UPDATE users SET subscription_active_until = $2, updated_at = $2 WHERE id = $1`
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
	const q = `
		SELECT id, email, password_hash, created_at, updated_at, subscription_active_until
		FROM users ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list users", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var u user.User
		var subUntil sql.NullTime
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt, &subUntil); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan user row", err)
		}
		if subUntil.Valid {
			u.SubscriptionActiveUntil = &subUntil.Time
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed while iterating users", err)
	}
	return users, nil
}
