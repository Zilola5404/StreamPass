package user

import (
	"context"
	"time"
)

// RefreshTokenID is the opaque identifier stored for a refresh token (a
// random string, never the JWT itself, so a leaked store dump can't be
// replayed as a token).
type RefreshTokenID string

// SessionStore is the port for refresh-token lifecycle management: issuing,
// validating and revoking. Backed by Redis in infrastructure, because
// sessions are short-lived, high-churn data — a poor fit for Postgres
// (spec section 12: Postgres + Redis).
type SessionStore interface {
	// Store persists a refresh token id for a user with a TTL matching the
	// token's expiry, so Redis itself expires stale sessions.
	Store(ctx context.Context, userID ID, tokenID RefreshTokenID, ttl time.Duration) error
	// IsValid reports whether tokenID is still an active session for userID.
	IsValid(ctx context.Context, userID ID, tokenID RefreshTokenID) (bool, error)
	// Revoke invalidates a single refresh token (used on logout).
	Revoke(ctx context.Context, userID ID, tokenID RefreshTokenID) error
	// RevokeAll invalidates every session for a user (used on password
	// change or suspected compromise — not exercised by MVP endpoints yet,
	// but the port is cheap to expose now and avoids a breaking interface
	// change later).
	RevokeAll(ctx context.Context, userID ID) error
}
