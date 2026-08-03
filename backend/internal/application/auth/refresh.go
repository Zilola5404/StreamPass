package auth

import (
	"context"
	"time"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

// RefreshUseCase implements "POST /refresh": mints a new access token when
// the client presents a valid, non-revoked refresh token.
type RefreshUseCase struct {
	tokens   TokenIssuer
	sessions user.SessionStore
	log      *logger.Logger
}

// NewRefreshUseCase wires the use case via constructor injection.
func NewRefreshUseCase(tokens TokenIssuer, sessions user.SessionStore, log *logger.Logger) *RefreshUseCase {
	return &RefreshUseCase{tokens: tokens, sessions: sessions, log: log.With("refresh")}
}

// AccessToken is the refresh response — only the short-lived credential
// rotates; the refresh token itself stays the same until logout.
type AccessToken struct {
	AccessToken     string
	AccessExpiresAt time.Time
}

// Execute validates the refresh token session and issues a fresh access token.
func (uc *RefreshUseCase) Execute(ctx context.Context, refreshToken string) (*AccessToken, error) {
	if refreshToken == "" {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "refresh_token must not be empty").
			WithDetails(map[string]any{"field": "refresh_token"})
	}

	userID, tokenID, err := uc.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeTokenInvalid, "invalid refresh token", err)
	}

	valid, err := uc.sessions.IsValid(ctx, userID, tokenID)
	if err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to validate session", err)
	}
	if !valid {
		return nil, apperrors.New(apperrors.CodeTokenInvalid, "refresh token revoked or expired")
	}

	access, accessExp, err := uc.tokens.IssueAccessToken(userID)
	if err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to issue access token", err)
	}

	return &AccessToken{
		AccessToken:     access,
		AccessExpiresAt: accessExp,
	}, nil
}
