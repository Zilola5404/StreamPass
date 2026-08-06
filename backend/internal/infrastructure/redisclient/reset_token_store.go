package redisclient

import (
	"context"
	"time"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
)

const resetTokenKeyPrefix = "pwdreset:"

// ResetTokenStore stores one-time password-reset tokens in Redis.
type ResetTokenStore struct {
	client *Client
}

// NewResetTokenStore builds a Redis-backed ResetTokenStore.
func NewResetTokenStore(client *Client) *ResetTokenStore {
	return &ResetTokenStore{client: client}
}

// Save stores token → userID with TTL.
func (s *ResetTokenStore) Save(ctx context.Context, token string, userID user.ID, ttl time.Duration) error {
	return s.client.Set(ctx, resetTokenKeyPrefix+token, string(userID), ttl)
}

// Consume returns the user ID and deletes the token (one-time use).
func (s *ResetTokenStore) Consume(ctx context.Context, token string) (user.ID, error) {
	key := resetTokenKeyPrefix + token
	val, err := s.client.Get(ctx, key)
	if err != nil {
		return "", err
	}
	_ = s.client.Del(ctx, key)
	if val == "" {
		return "", apperrors.New(apperrors.CodeNotFound, "reset token not found")
	}
	return user.ID(val), nil
}
