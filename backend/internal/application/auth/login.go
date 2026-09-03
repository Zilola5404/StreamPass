package auth

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"streampass/backend/internal/domain/subscription"
	"streampass/backend/internal/domain/user"
	"streampass/shared/idgen"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

// LoginInput carries credentials and optional device registration fields.
type LoginInput struct {
	Email      string
	Password   string
	DeviceID   string
	DeviceName string
}

// LoginUseCase implements "POST /login" (spec section 13/22).
type LoginUseCase struct {
	repo        user.Repository
	devices     user.DeviceRepository
	hasher      PasswordHasher
	tokens      TokenIssuer
	sessions    user.SessionStore
	clock       Clock
	maxDevices  int
	log         *logger.Logger
}

// NewLoginUseCase wires the use case via constructor injection.
func NewLoginUseCase(
	repo user.Repository,
	devices user.DeviceRepository,
	hasher PasswordHasher,
	tokens TokenIssuer,
	sessions user.SessionStore,
	clock Clock,
	maxDevices int,
	log *logger.Logger,
) *LoginUseCase {
	if maxDevices <= 0 {
		maxDevices = 2
	}
	return &LoginUseCase{
		repo: repo, devices: devices, hasher: hasher, tokens: tokens,
		sessions: sessions, clock: clock, maxDevices: maxDevices, log: log.With("login"),
	}
}

// Execute verifies credentials and, on success, issues a fresh token pair
// and records the refresh token as an active session in Redis.
// When DeviceID is set, the device is registered / refreshed and the
// multi-device limit is enforced (BL-049).
func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (*TokenPair, error) {
	u, err := uc.repo.FindByEmail(ctx, in.Email)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == apperrors.CodeNotFound {
			return nil, apperrors.New(apperrors.CodeInvalidCredentials, "invalid email or password")
		}
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to look up user", err)
	}

	if !uc.hasher.Verify(u.PasswordHash, in.Password) {
		return nil, apperrors.New(apperrors.CodeInvalidCredentials, "invalid email or password")
	}

	if u.IsBanned() {
		return nil, apperrors.New(apperrors.CodeForbidden, "аккаунт заблокирован")
	}

	deviceID := strings.TrimSpace(in.DeviceID)
	deviceName := strings.TrimSpace(in.DeviceName)
	var existing *user.Device
	if deviceID != "" {
		if utf8.RuneCountInString(deviceID) > 128 {
			return nil, apperrors.New(apperrors.CodeInvalidInput, "device_id too long")
		}
		if utf8.RuneCountInString(deviceName) > 64 {
			return nil, apperrors.New(apperrors.CodeInvalidInput, "device_name too long")
		}
		if deviceName == "" {
			deviceName = "Устройство"
		}
		existing, err = uc.devices.FindByUserAndDevice(ctx, u.ID, user.DeviceID(deviceID))
		if err != nil {
			uc.log.Error(ctx, err)
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to look up device", err)
		}
		if existing == nil {
			n, err := uc.devices.CountByUser(ctx, u.ID)
			if err != nil {
				uc.log.Error(ctx, err)
				return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to count devices", err)
			}
			limit := uc.maxDevices
			if planLimit := subscription.MaxDevicesForPlan(u.PlanCode); planLimit > 0 {
				limit = planLimit
			} else if u.EntitlementSource == "trial" {
				limit = subscription.MaxDevicesForPlan(subscription.PlanPersonalBasic)
			}
			if n >= limit {
				return nil, apperrors.New(
					apperrors.CodeDeviceLimit,
					"Достигнут лимит устройств. Отключите одно в профиле.",
				).WithDetails(map[string]any{"max_devices": limit})
			}
		}
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

	if deviceID != "" {
		now := uc.clock.Now()
		row := &user.Device{
			UserID:         u.ID,
			DeviceID:       user.DeviceID(deviceID),
			Name:           deviceName,
			RefreshTokenID: refreshID,
			LastSeenAt:     now,
		}
		if existing != nil {
			if existing.RefreshTokenID != "" && existing.RefreshTokenID != refreshID {
				_ = uc.sessions.Revoke(ctx, u.ID, existing.RefreshTokenID)
			}
			row.ID = existing.ID
			row.CreatedAt = existing.CreatedAt
			if err := uc.devices.UpdateSession(ctx, row); err != nil {
				uc.log.Error(ctx, err)
				return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to update device", err)
			}
		} else {
			row.ID = user.DeviceRowID(idgen.New())
			row.CreatedAt = now
			if err := uc.devices.Create(ctx, row); err != nil {
				uc.log.Error(ctx, err)
				return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to register device", err)
			}
		}
	}

	return &TokenPair{
		AccessToken:      access,
		AccessExpiresAt:  accessExp,
		RefreshToken:     refresh,
		RefreshExpiresAt: refreshExp,
	}, nil
}
