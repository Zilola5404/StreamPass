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
	// Billing module; Auth only reads it, never writes it (except trial on register).
	SubscriptionActiveUntil *time.Time
	// TrialStartedAt / TrialEndsAt are set once on registration (architect trial model).
	TrialStartedAt *time.Time
	TrialEndsAt    *time.Time
	// EntitlementSource: trial | paid | admin (empty = legacy paid/until only).
	EntitlementSource string
	// PlanCode is the last confirmed product plan (personal_basic | personal_pro | business).
	PlanCode string
	// SubscriptionCanceledAt marks auto-renew cancel; access continues until ActiveUntil.
	SubscriptionCanceledAt *time.Time
	// BannedAt is set by admin ban (BL-050); nil means not banned.
	BannedAt *time.Time
}

// IsSubscriptionActive reports whether the user currently has access,
// per spec section 22 ("Оплата и активация подписки").
func (u *User) IsSubscriptionActive(now time.Time) bool {
	if u.IsBanned() {
		return false
	}
	return u.SubscriptionActiveUntil != nil && u.SubscriptionActiveUntil.After(now)
}

// IsBanned reports whether the account is admin-banned.
func (u *User) IsBanned() bool {
	return u.BannedAt != nil
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
	// ActivatePaidPlan extends access after a confirmed payment and records plan_code.
	ActivatePaidPlan(ctx context.Context, id ID, activeUntil time.Time, planCode string) error
	// CancelAutoRenew sets subscription_canceled_at (access until ActiveUntil).
	CancelAutoRenew(ctx context.Context, id ID, now time.Time) error
	// ClearSubscription removes Premium (admin revoke / ban).
	ClearSubscription(ctx context.Context, id ID, now time.Time) error
	// SetBanned marks or clears the ban timestamp.
	SetBanned(ctx context.Context, id ID, bannedAt *time.Time, now time.Time) error
	// SearchByEmail returns users whose email contains q (case-insensitive).
	// Empty q lists all users (same as List).
	SearchByEmail(ctx context.Context, q string) ([]*User, error)
	// List returns every registered user, newest first. Used by the admin
	// user-listing endpoint — there is no pagination yet (fine at MVP
	// scale; revisit if the user table grows large enough for this to
	// matter, per YAGNI).
	List(ctx context.Context) ([]*User, error)
	// UpdatePasswordHash replaces the stored Argon2id hash (password change
	// / reset). Callers must revoke sessions separately.
	UpdatePasswordHash(ctx context.Context, id ID, passwordHash string, now time.Time) error
	// Delete permanently removes the user and dependent payment rows.
	Delete(ctx context.Context, id ID) error
}

// ErrNotFound is a sentinel-style helper so infrastructure implementations
// return a consistent AppError for "no such user".
func ErrNotFound(email string) error {
	return apperrors.New(apperrors.CodeNotFound, "user not found").WithDetails(map[string]any{"email": email})
}
