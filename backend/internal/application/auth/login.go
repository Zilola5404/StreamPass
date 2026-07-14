package auth

import (
	"context"
	"errors"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

// LoginUseCase implements "POST /login" (spec section 13/22).
type LoginUseCase struct {
	repo     user.Repository
	hasher   PasswordHasher
	tokens   TokenIssuer
	sessions user.SessionStore
	clock    Clock
	log      *logger.Logger
}

// NewLoginUseCase wires the use case via constructor injection.
func NewLoginUseCase(repo user.Repository, hasher PasswordHasher, tokens TokenIssuer, sessions user.SessionStore, clock Clock, log *logger.Logger) *LoginUseCase {
	return &LoginUseCase{repo: repo, hasher: hasher, tokens: tokens, sessions: sessions, clock: clock, log: log.With("login")}
}

// Execute verifies credentials and, on success, issues a fresh token pair
// and records the refresh token as an active session in Redis.
func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*TokenPair, error) {
	u, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == apperrors.CodeNotFound {
			return nil, apperrors.New(apperrors.CodeInvalidCredentials, "invalid email or password")
		}
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to look up user", err)
	}

	if !uc.hasher.Verify(u.PasswordHash, password) {
		return nil, apperrors.New(apperrors.CodeInvalidCredentials, "invalid email or password")
	}

	access, accessExp, err := uc.tokens.IssueAccessToken(u.ID)
	if err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to issue access token", err)
	}
	refresh, refreshID, refreshExp, err := uc.tokens.IssueRefreshToken(u.ID)
	if err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to issue refresh token", err)
	}

	sessionTTL := refreshExp.Sub(uc.clock.Now())
	if err := uc.sessions.Store(ctx, u.ID, refreshID, sessionTTL); err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to persist session", err)
	}

	return &TokenPair{
		AccessToken:      access,
		AccessExpiresAt:  accessExp,
		RefreshToken:     refresh,
		RefreshExpiresAt: refreshExp,
	}, nil
}
