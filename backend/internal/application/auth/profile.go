package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/idgen"
	"streampass/shared/logger"
)

// ResetTokenTTL is how long a password-reset token remains valid.
const ResetTokenTTL = time.Hour

// ResetTokenStore persists one-time password-reset tokens (Redis in prod).
type ResetTokenStore interface {
	Save(ctx context.Context, token string, userID user.ID, ttl time.Duration) error
	Consume(ctx context.Context, token string) (user.ID, error)
}

// Profile is the read model for GET /me.
type Profile struct {
	Email                   string
	CreatedAt               time.Time
	SubscriptionActiveUntil *time.Time
}

// GetProfileUseCase implements GET /me.
type GetProfileUseCase struct {
	repo user.Repository
	log  *logger.Logger
}

// NewGetProfileUseCase wires the use case.
func NewGetProfileUseCase(repo user.Repository, log *logger.Logger) *GetProfileUseCase {
	return &GetProfileUseCase{repo: repo, log: log.With("get_profile")}
}

// Execute returns the authenticated user's profile.
func (uc *GetProfileUseCase) Execute(ctx context.Context, userID user.ID) (*Profile, error) {
	u, err := uc.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &Profile{
		Email:                   u.Email,
		CreatedAt:               u.CreatedAt,
		SubscriptionActiveUntil: u.SubscriptionActiveUntil,
	}, nil
}

// ChangePasswordUseCase implements PUT /me/password.
type ChangePasswordUseCase struct {
	repo     user.Repository
	hasher   PasswordHasher
	sessions user.SessionStore
	clock    Clock
	log      *logger.Logger
}

// NewChangePasswordUseCase wires the use case.
func NewChangePasswordUseCase(repo user.Repository, hasher PasswordHasher, sessions user.SessionStore, clock Clock, log *logger.Logger) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{repo: repo, hasher: hasher, sessions: sessions, clock: clock, log: log.With("change_password")}
}

// Execute verifies the current password, sets a new hash, and revokes all sessions.
func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID user.ID, currentPassword, newPassword string) error {
	if len(newPassword) < minPasswordLength {
		return apperrors.New(apperrors.CodeInvalidInput, "password too short").
			WithDetails(map[string]any{"field": "new_password", "min_length": minPasswordLength})
	}
	u, err := uc.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !uc.hasher.Verify(u.PasswordHash, currentPassword) {
		return apperrors.New(apperrors.CodeUnauthorized, "current password is incorrect").
			WithDetails(map[string]any{"field": "current_password"})
	}
	hash, err := uc.hasher.Hash(newPassword)
	if err != nil {
		uc.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to hash password", err)
	}
	if err := uc.repo.UpdatePasswordHash(ctx, userID, hash, uc.clock.Now()); err != nil {
		uc.log.Error(ctx, err)
		return err
	}
	if err := uc.sessions.RevokeAll(ctx, userID); err != nil {
		uc.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to revoke sessions", err)
	}
	return nil
}

// DeleteAccountUseCase implements DELETE /me.
type DeleteAccountUseCase struct {
	repo     user.Repository
	sessions user.SessionStore
	log      *logger.Logger
}

// NewDeleteAccountUseCase wires the use case.
func NewDeleteAccountUseCase(repo user.Repository, sessions user.SessionStore, log *logger.Logger) *DeleteAccountUseCase {
	return &DeleteAccountUseCase{repo: repo, sessions: sessions, log: log.With("delete_account")}
}

// Execute permanently deletes the account and revokes sessions.
func (uc *DeleteAccountUseCase) Execute(ctx context.Context, userID user.ID) error {
	_ = uc.sessions.RevokeAll(ctx, userID)
	if err := uc.repo.Delete(ctx, userID); err != nil {
		uc.log.Error(ctx, err)
		return err
	}
	return nil
}

// ForgotPasswordResult is returned to the HTTP layer. ResetToken is only
// populated when the server is configured to expose it (no SMTP yet).
type ForgotPasswordResult struct {
	Message    string
	ResetToken string // empty unless ExposeResetToken
}

// ForgotPasswordUseCase implements POST /password/forgot.
type ForgotPasswordUseCase struct {
	repo             user.Repository
	tokens           ResetTokenStore
	exposeResetToken bool
	log              *logger.Logger
}

// NewForgotPasswordUseCase wires the use case.
func NewForgotPasswordUseCase(repo user.Repository, tokens ResetTokenStore, exposeResetToken bool, log *logger.Logger) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{repo: repo, tokens: tokens, exposeResetToken: exposeResetToken, log: log.With("forgot_password")}
}

// Execute always returns a generic success message (anti-enumeration).
func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, email string) (*ForgotPasswordResult, error) {
	const msg = "Если аккаунт с таким адресом существует, инструкции отправлены."
	email = strings.TrimSpace(email)
	if !emailPattern.MatchString(email) {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "invalid email format").
			WithDetails(map[string]any{"field": "email"})
	}

	result := &ForgotPasswordResult{Message: msg}
	u, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == apperrors.CodeNotFound {
			return result, nil
		}
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to look up user", err)
	}

	token := idgen.New()
	if err := uc.tokens.Save(ctx, token, u.ID, ResetTokenTTL); err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to store reset token", err)
	}
	if uc.exposeResetToken {
		result.ResetToken = token
		uc.log.Info(ctx, "password reset token issued (expose_reset_token=true)")
	}
	return result, nil
}

// ResetPasswordUseCase implements POST /password/reset.
type ResetPasswordUseCase struct {
	repo     user.Repository
	hasher   PasswordHasher
	tokens   ResetTokenStore
	sessions user.SessionStore
	clock    Clock
	log      *logger.Logger
}

// NewResetPasswordUseCase wires the use case.
func NewResetPasswordUseCase(repo user.Repository, hasher PasswordHasher, tokens ResetTokenStore, sessions user.SessionStore, clock Clock, log *logger.Logger) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{repo: repo, hasher: hasher, tokens: tokens, sessions: sessions, clock: clock, log: log.With("reset_password")}
}

// Execute consumes the reset token and sets a new password.
func (uc *ResetPasswordUseCase) Execute(ctx context.Context, token, newPassword string) error {
	if strings.TrimSpace(token) == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "reset token required").
			WithDetails(map[string]any{"field": "token"})
	}
	if len(newPassword) < minPasswordLength {
		return apperrors.New(apperrors.CodeInvalidInput, "password too short").
			WithDetails(map[string]any{"field": "password", "min_length": minPasswordLength})
	}

	userID, err := uc.tokens.Consume(ctx, token)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == apperrors.CodeNotFound {
			return apperrors.New(apperrors.CodeUnauthorized, "invalid or expired reset token")
		}
		uc.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to consume reset token", err)
	}

	hash, err := uc.hasher.Hash(newPassword)
	if err != nil {
		uc.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to hash password", err)
	}
	if err := uc.repo.UpdatePasswordHash(ctx, userID, hash, uc.clock.Now()); err != nil {
		uc.log.Error(ctx, err)
		return err
	}
	_ = uc.sessions.RevokeAll(ctx, userID)
	return nil
}
