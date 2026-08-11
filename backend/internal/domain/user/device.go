package user

import (
	"context"
	"time"
)

// DeviceID is the client-stable identifier for a physical/logical device
// (generated once on the client and reused across logins).
type DeviceID string

// DeviceRowID is the server-side primary key for a registered device row.
type DeviceRowID string

// Device is a registered client for multi-device limit / revoke (BL-049).
type Device struct {
	ID             DeviceRowID
	UserID         ID
	DeviceID       DeviceID
	Name           string
	RefreshTokenID RefreshTokenID
	CreatedAt      time.Time
	LastSeenAt     time.Time
}

// DeviceRepository persists registered devices for a user.
type DeviceRepository interface {
	FindByUserAndDevice(ctx context.Context, userID ID, deviceID DeviceID) (*Device, error)
	FindByID(ctx context.Context, id DeviceRowID) (*Device, error)
	ListByUser(ctx context.Context, userID ID) ([]*Device, error)
	CountByUser(ctx context.Context, userID ID) (int, error)
	Create(ctx context.Context, d *Device) error
	UpdateSession(ctx context.Context, d *Device) error
	Delete(ctx context.Context, id DeviceRowID) error
}
