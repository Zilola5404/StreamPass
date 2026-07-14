package auth

import (
	"context"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

// LogoutUseCase implements "POST /logout" (spec section 13): revokes the
// refresh token's session so it can no longer be used to mint new access
// tokens.
type LogoutUseCase struct {
	tokens   TokenIssuer
	sessions user.SessionStore
	log      *logger.Logger
}

// NewLogoutUseCase wires the use case via constructor injection.
func NewLogoutUseCase(tokens TokenIssuer, sessions user.SessionStore, log *logger.Logger) *LogoutUseCase {
	return &LogoutUseCase{tokens: tokens, sessions: sessions, log: log.With("logout")}
}

// Execute parses the refresh token and revokes the matching session.
func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	userID, tokenID, err := uc.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeTokenInvalid, "invalid refresh token", err)
	}
	if err := uc.sessions.Revoke(ctx, userID, tokenID); err != nil {
		uc.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to revoke session", err)
	}
	return nil
}
