package auth

import (
	"context"
	"time"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

// DeviceInfo is a device row exposed to the profile UI (BL-049).
type DeviceInfo struct {
	ID         string
	DeviceID   string
	Name       string
	CreatedAt  time.Time
	LastSeenAt time.Time
}

// DeviceList is the GET /me/devices payload.
type DeviceList struct {
	Devices    []DeviceInfo
	MaxDevices int
}

// ListDevicesUseCase implements GET /me/devices.
type ListDevicesUseCase struct {
	devices    user.DeviceRepository
	maxDevices int
	log        *logger.Logger
}

// NewListDevicesUseCase wires the use case.
func NewListDevicesUseCase(devices user.DeviceRepository, maxDevices int, log *logger.Logger) *ListDevicesUseCase {
	if maxDevices <= 0 {
		maxDevices = 3
	}
	return &ListDevicesUseCase{devices: devices, maxDevices: maxDevices, log: log.With("list_devices")}
}

// Execute returns registered devices for the user.
func (uc *ListDevicesUseCase) Execute(ctx context.Context, userID user.ID) (*DeviceList, error) {
	rows, err := uc.devices.ListByUser(ctx, userID)
	if err != nil {
		uc.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list devices", err)
	}
	out := make([]DeviceInfo, 0, len(rows))
	for _, d := range rows {
		out = append(out, DeviceInfo{
			ID:         string(d.ID),
			DeviceID:   string(d.DeviceID),
			Name:       d.Name,
			CreatedAt:  d.CreatedAt,
			LastSeenAt: d.LastSeenAt,
		})
	}
	return &DeviceList{Devices: out, MaxDevices: uc.maxDevices}, nil
}

// RevokeDeviceUseCase implements DELETE /me/devices/{id}.
type RevokeDeviceUseCase struct {
	devices  user.DeviceRepository
	sessions user.SessionStore
	log      *logger.Logger
}

// NewRevokeDeviceUseCase wires the use case.
func NewRevokeDeviceUseCase(devices user.DeviceRepository, sessions user.SessionStore, log *logger.Logger) *RevokeDeviceUseCase {
	return &RevokeDeviceUseCase{devices: devices, sessions: sessions, log: log.With("revoke_device")}
}

// Execute removes the device and revokes its refresh session when present.
func (uc *RevokeDeviceUseCase) Execute(ctx context.Context, userID user.ID, deviceRowID string) error {
	d, err := uc.devices.FindByID(ctx, user.DeviceRowID(deviceRowID))
	if err != nil {
		return err
	}
	if d.UserID != userID {
		return apperrors.New(apperrors.CodeNotFound, "device not found")
	}
	if d.RefreshTokenID != "" {
		_ = uc.sessions.Revoke(ctx, userID, d.RefreshTokenID)
	}
	if err := uc.devices.Delete(ctx, d.ID); err != nil {
		uc.log.Error(ctx, err)
		return err
	}
	return nil
}
