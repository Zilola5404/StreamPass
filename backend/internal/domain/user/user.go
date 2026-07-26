// Package user contains the Auth bounded context's domain model: the User
// aggregate and the port (interface) infrastructure must implement to
// persist it. No infrastructure or transport concern belongs here — pure
// business rules only (DDD, Clean Architecture: Domain layer).
package user

import (
	"context"
	"time"

	apperrors "streampass/shared/errors"
)

// ID uniquely identifies a User.
type ID string

// User is the Auth aggregate root. PasswordHash is always an Argon2id hash
// (spec section 17, Security) — plaintext passwords never enter this type.
type User struct {
	ID           ID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// SubscriptionActiveUntil is a denormalized read model updated by the
	// Billing module; Auth only reads it, never writes it.
	SubscriptionActiveUntil *time.Time
}

// IsSubscriptionActive reports whether the user currently has access,
// per spec section 22 ("Оплата и активация подписки").
func (u *User) IsSubscriptionActive(now time.Time) bool {
	return u.SubscriptionActiveUntil != nil && u.SubscriptionActiveUntil.After(now)
}

// NewUser constructs a new User aggregate with a freshly hashed password.
// Validation of the raw email/password happens in the application layer
// (use case), since it depends on policy (min length etc.) that may evolve
// independently of the entity shape.
func NewUser(id ID, email, passwordHash string, now time.Time) *User {
	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Repository is the port the Auth use cases depend on. Infrastructure
// (Postgres) implements this interface; the application layer never knows
// about SQL (Dependency Injection / Interface First).
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id ID) (*User, error)
	// ExtendSubscription sets SubscriptionActiveUntil for a user. Owned by
	// this port (not a separate subscription.Repository) because it's a
	// single denormalized field on the User row — see
	// domain/subscription's package doc for the reasoning.
	ExtendSubscription(ctx context.Context, id ID, activeUntil time.Time) error
	// List returns every registered user, newest first. Used by the admin
	// user-listing endpoint — there is no pagination yet (fine at MVP
	// scale; revisit if the user table grows large enough for this to
	// matter, per YAGNI).
	List(ctx context.Context) ([]*User, error)
}

// ErrNotFound is a sentinel-style helper so infrastructure implementations
// return a consistent AppError for "no such user".
func ErrNotFound(email string) error {
	return apperrors.New(apperrors.CodeNotFound, "user not found").WithDetails(map[string]any{"email": email})
}
