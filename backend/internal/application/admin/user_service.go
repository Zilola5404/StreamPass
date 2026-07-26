package admin

import (
	"context"
	"time"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type UserSummary struct {
	ID                      user.ID
	Email                   string
	CreatedAt               time.Time
	SubscriptionActiveUntil *time.Time
	IsSubscriptionActive    bool
}

type UserService struct {
	repo  user.Repository
	clock Clock
	log   *logger.Logger
}

func NewUserService(repo user.Repository, clock Clock, log *logger.Logger) *UserService {
	return &UserService{repo: repo, clock: clock, log: log.With("admin_user_service")}
}

func (s *UserService) ListUsers(ctx context.Context) ([]UserSummary, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list users", err)
	}

	now := s.clock.Now()
	summaries := make([]UserSummary, len(users))
	for i, u := range users {
		summaries[i] = UserSummary{
			ID:                      u.ID,
			Email:                   u.Email,
			CreatedAt:               u.CreatedAt,
			SubscriptionActiveUntil: u.SubscriptionActiveUntil,
			IsSubscriptionActive:    u.IsSubscriptionActive(now),
		}
	}
	return summaries, nil
}
