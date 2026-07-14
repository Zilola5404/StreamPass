package redisclient

import (
	"context"
	"errors"
	"time"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
)

// sessionKeyPrefix namespaces refresh-token session keys in the shared
// Redis instance. Format: session:{userID}:{tokenID} -> "1".
const sessionKeyPrefix = "session"

// SessionStore implements user.SessionStore on top of the minimal Redis
// client. Each active refresh token is one key with a TTL matching its
// expiry, so Redis itself garbage-collects stale sessions.
type SessionStore struct {
	client *Client
}

// NewSessionStore builds a Redis-backed user.SessionStore.
func NewSessionStore(client *Client) *SessionStore {
	return &SessionStore{client: client}
}

func sessionKey(userID user.ID, tokenID user.RefreshTokenID) string {
	return sessionKeyPrefix + ":" + string(userID) + ":" + string(tokenID)
}

func sessionUserPrefix(userID user.ID) string {
	return sessionKeyPrefix + ":" + string(userID) + ":"
}

// Store persists a refresh token id for a user with a TTL.
func (s *SessionStore) Store(ctx context.Context, userID user.ID, tokenID user.RefreshTokenID, ttl time.Duration) error {
	return s.client.Set(ctx, sessionKey(userID, tokenID), "1", ttl)
}

// IsValid reports whether tokenID is still an active session for userID.
func (s *SessionStore) IsValid(ctx context.Context, userID user.ID, tokenID user.RefreshTokenID) (bool, error) {
	return s.client.Exists(ctx, sessionKey(userID, tokenID))
}

// Revoke invalidates a single refresh token.
func (s *SessionStore) Revoke(ctx context.Context, userID user.ID, tokenID user.RefreshTokenID) error {
	if err := s.client.Del(ctx, sessionKey(userID, tokenID)); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return err
		}
		return apperrors.Wrap(apperrors.CodeInternal, "failed to revoke session", err)
	}
	return nil
}

// RevokeAll invalidates every session for a user.
func (s *SessionStore) RevokeAll(ctx context.Context, userID user.ID) error {
	return s.client.DelPrefix(ctx, sessionUserPrefix(userID))
}
