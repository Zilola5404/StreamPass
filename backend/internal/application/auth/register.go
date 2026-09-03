package auth

import (
	"context"
	"errors"
	"time"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

// IDGenerator is the port for generating new user IDs, injected so the
// use case doesn't hard-depend on a specific UUID library/version.
type IDGenerator interface {
	NewID() user.ID
}

// RegisterUseCase implements "POST /register" business logic (spec section
// 13/22: user registration).
type RegisterUseCase struct {
	repo       user.Repository
	hasher     PasswordHasher
	ids        IDGenerator
	clock      Clock
	log        *logger.Logger
	trialHours int
}

// DefaultTrialHours is the free-trial length in wall-clock hours (BILLING-002 SSOT).
// Config key: billing.trial_hours (default 72). Never trust client/device clock.
const DefaultTrialHours = 72

// NewRegisterUseCase wires the use case via constructor injection — every
// dependency is an interface (Dependency Injection / Interface First).
func NewRegisterUseCase(repo user.Repository, hasher PasswordHasher, ids IDGenerator, clock Clock, log *logger.Logger) *RegisterUseCase {
	return &RegisterUseCase{
		repo: repo, hasher: hasher, ids: ids, clock: clock,
		log: log.With("register"), trialHours: DefaultTrialHours,
	}
}

// WithTrialHours overrides the free-trial length in hours (tests / billing.trial_hours).
func (uc *RegisterUseCase) WithTrialHours(hours int) *RegisterUseCase {
	if hours > 0 {
		uc.trialHours = hours
	}
	return uc
}

// WithTrialDays is a test helper: days×24 hours. Production uses WithTrialHours / billing.trial_hours.
func (uc *RegisterUseCase) WithTrialDays(days int) *RegisterUseCase {
	if days > 0 {
		uc.trialHours = days * 24
	}
	return uc
}

// Execute validates input, ensures the email is not taken, hashes the
// password and persists a new User with a server-side free trial.
func (uc *RegisterUseCase) Execute(ctx context.Context, email, password string) (*user.User, error) {
	if err := validateCredentials(email, password); err != nil {
		return nil, err
	}

	if _, err := uc.repo.FindByEmail(ctx, email); err == nil {
		return nil, apperrors.New(apperrors.CodeAlreadyExists, "email already registered").
			WithDetails(map[string]any{"field": "email"})
	} else {
		var appErr *apperrors.AppError
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeNotFound {
			uc.log.Error(ctx, err)
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to check existing user", err)
		}
	}

	hash, err := uc.hasher.Hash(password)
	if err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to hash password", err)
	}

	now := uc.clock.Now()
	u := user.NewUser(uc.ids.NewID(), email, hash, now)
	trialEnd := now.Add(time.Duration(uc.trialHours) * time.Hour)
	u.TrialStartedAt = &now
	u.TrialEndsAt = &trialEnd
	u.SubscriptionActiveUntil = &trialEnd
	u.EntitlementSource = "trial"

	if err := uc.repo.Create(ctx, u); err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to create user", err)
	}

	return u, nil
}
