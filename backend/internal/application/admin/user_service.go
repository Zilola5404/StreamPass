package admin

import (
	"context"
	"strings"
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
	Banned                  bool
	BannedAt                *time.Time
}

type AuditSummary struct {
	ID         string
	Actor      string
	Action     string
	TargetType string
	TargetID   string
	Detail     map[string]any
	CreatedAt  time.Time
}

type UserService struct {
	repo     user.Repository
	sessions user.SessionStore
	audit    user.AuditRepository
	clock    Clock
	log      *logger.Logger
}

func NewUserService(
	repo user.Repository,
	sessions user.SessionStore,
	audit user.AuditRepository,
	clock Clock,
	log *logger.Logger,
) *UserService {
	return &UserService{
		repo: repo, sessions: sessions, audit: audit, clock: clock,
		log: log.With("admin_user_service"),
	}
}

func (s *UserService) ListUsers(ctx context.Context, query string) ([]UserSummary, error) {
	var users []*user.User
	var err error
	if strings.TrimSpace(query) == "" {
		users, err = s.repo.List(ctx)
	} else {
		users, err = s.repo.SearchByEmail(ctx, query)
	}
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
			Banned:                  u.IsBanned(),
			BannedAt:                u.BannedAt,
		}
	}
	return summaries, nil
}

// GrantPremium extends subscription by days from max(now, current expiry).
func (s *UserService) GrantPremium(ctx context.Context, actor, userID string, days int) (*UserSummary, error) {
	if days <= 0 || days > 3650 {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "days must be between 1 and 3650")
	}
	u, err := s.repo.FindByID(ctx, user.ID(userID))
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	base := now
	if u.SubscriptionActiveUntil != nil && u.SubscriptionActiveUntil.After(now) {
		base = *u.SubscriptionActiveUntil
	}
	until := base.Add(time.Duration(days) * 24 * time.Hour)
	if err := s.repo.ExtendSubscription(ctx, u.ID, until); err != nil {
		s.log.Error(ctx, err)
		return nil, err
	}
	_ = s.audit.Append(ctx, &user.AuditEntry{
		Actor: actor, Action: "grant_premium", TargetType: "user", TargetID: string(u.ID),
		Detail: map[string]any{"email": u.Email, "days": days, "until": until.Format(time.RFC3339)},
		CreatedAt: now,
	})
	u.SubscriptionActiveUntil = &until
	return &UserSummary{
		ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt,
		SubscriptionActiveUntil: &until, IsSubscriptionActive: true,
		Banned: u.IsBanned(), BannedAt: u.BannedAt,
	}, nil
}

// RevokePremium clears subscription_active_until.
func (s *UserService) RevokePremium(ctx context.Context, actor, userID string) (*UserSummary, error) {
	u, err := s.repo.FindByID(ctx, user.ID(userID))
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := s.repo.ClearSubscription(ctx, u.ID, now); err != nil {
		s.log.Error(ctx, err)
		return nil, err
	}
	_ = s.audit.Append(ctx, &user.AuditEntry{
		Actor: actor, Action: "revoke_premium", TargetType: "user", TargetID: string(u.ID),
		Detail: map[string]any{"email": u.Email}, CreatedAt: now,
	})
	return &UserSummary{
		ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt,
		IsSubscriptionActive: false, Banned: u.IsBanned(), BannedAt: u.BannedAt,
	}, nil
}

// BanUser bans the account, clears Premium, and revokes sessions.
func (s *UserService) BanUser(ctx context.Context, actor, userID string) (*UserSummary, error) {
	u, err := s.repo.FindByID(ctx, user.ID(userID))
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := s.repo.ClearSubscription(ctx, u.ID, now); err != nil {
		s.log.Error(ctx, err)
		return nil, err
	}
	if err := s.repo.SetBanned(ctx, u.ID, &now, now); err != nil {
		s.log.Error(ctx, err)
		return nil, err
	}
	if s.sessions != nil {
		_ = s.sessions.RevokeAll(ctx, u.ID)
	}
	_ = s.audit.Append(ctx, &user.AuditEntry{
		Actor: actor, Action: "ban_user", TargetType: "user", TargetID: string(u.ID),
		Detail: map[string]any{"email": u.Email}, CreatedAt: now,
	})
	return &UserSummary{
		ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt,
		IsSubscriptionActive: false, Banned: true, BannedAt: &now,
	}, nil
}

// UnbanUser clears the ban flag.
func (s *UserService) UnbanUser(ctx context.Context, actor, userID string) (*UserSummary, error) {
	u, err := s.repo.FindByID(ctx, user.ID(userID))
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := s.repo.SetBanned(ctx, u.ID, nil, now); err != nil {
		s.log.Error(ctx, err)
		return nil, err
	}
	_ = s.audit.Append(ctx, &user.AuditEntry{
		Actor: actor, Action: "unban_user", TargetType: "user", TargetID: string(u.ID),
		Detail: map[string]any{"email": u.Email}, CreatedAt: now,
	})
	u.BannedAt = nil
	return &UserSummary{
		ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt,
		SubscriptionActiveUntil: u.SubscriptionActiveUntil,
		IsSubscriptionActive:    u.IsSubscriptionActive(now),
		Banned:                  false,
	}, nil
}

// ListAudit returns recent admin audit events.
func (s *UserService) ListAudit(ctx context.Context, limit int) ([]AuditSummary, error) {
	rows, err := s.audit.List(ctx, limit)
	if err != nil {
		s.log.Error(ctx, err)
		return nil, err
	}
	out := make([]AuditSummary, 0, len(rows))
	for _, e := range rows {
		out = append(out, AuditSummary{
			ID: e.ID, Actor: e.Actor, Action: e.Action,
			TargetType: e.TargetType, TargetID: e.TargetID,
			Detail: e.Detail, CreatedAt: e.CreatedAt,
		})
	}
	return out, nil
}

// RecordAudit appends a free-form admin action (rules publish, relay delete…).
func (s *UserService) RecordAudit(ctx context.Context, actor, action, targetType, targetID string, detail map[string]any) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Append(ctx, &user.AuditEntry{
		Actor: actor, Action: action, TargetType: targetType, TargetID: targetID,
		Detail: detail, CreatedAt: s.clock.Now(),
	})
}
