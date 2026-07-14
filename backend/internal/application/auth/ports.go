// Package auth contains the Auth module's application layer: use cases that
// orchestrate the user.User domain model against ports implemented by
// infrastructure. No SQL, no HTTP, no JWT library specifics live here —
// only interfaces and orchestration (Clean Architecture: Application layer).
package auth

import (
	"time"

	"streampass/backend/internal/domain/user"
)

// PasswordHasher is the port for one-way password hashing. Infrastructure
// implements it with Argon2id per spec section 17 (Security).
type PasswordHasher interface {
	Hash(plaintext string) (string, error)
	Verify(hash, plaintext string) bool
}

// TokenPair is the pair of tokens returned to a client after a successful
// login: a short-lived JWT access token and a long-lived opaque-ish refresh
// token.
type TokenPair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

// TokenIssuer is the port for JWT issuance/verification. Infrastructure
// implements it (spec section 17: JWT).
type TokenIssuer interface {
	IssueAccessToken(userID user.ID) (token string, expiresAt time.Time, err error)
	IssueRefreshToken(userID user.ID) (token string, id user.RefreshTokenID, expiresAt time.Time, err error)
	ParseAccessToken(token string) (user.ID, error)
	ParseRefreshToken(token string) (user.ID, user.RefreshTokenID, error)
}

// Clock is injected instead of calling time.Now() directly, so use cases
// are deterministic in tests.
type Clock interface {
	Now() time.Time
}

// SystemClock is the production Clock implementation.
type SystemClock struct{}

// Now returns the current wall-clock time.
func (SystemClock) Now() time.Time { return time.Now().UTC() }
