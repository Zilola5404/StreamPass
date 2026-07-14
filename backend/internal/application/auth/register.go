package auth

import (
	"context"
	"errors"

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
	repo   user.Repository
	hasher PasswordHasher
	ids    IDGenerator
	clock  Clock
	log    *logger.Logger
}

// NewRegisterUseCase wires the use case via constructor injection — every
// dependency is an interface (Dependency Injection / Interface First).
func NewRegisterUseCase(repo user.Repository, hasher PasswordHasher, ids IDGenerator, clock Clock, log *logger.Logger) *RegisterUseCase {
	return &RegisterUseCase{repo: repo, hasher: hasher, ids: ids, clock: clock, log: log.With("register")}
}

// Execute validates input, ensures the email is not taken, hashes the
// password and persists a new User.
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

	u := user.NewUser(uc.ids.NewID(), email, hash, uc.clock.Now())
	if err := uc.repo.Create(ctx, u); err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to create user", err)
	}

	return u, nil
}
