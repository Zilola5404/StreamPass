package postgres

import (
	"context"
	"database/sql"
	"errors"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
)

// DeviceRepository stores registered client devices (BL-049).
type DeviceRepository struct {
	db *sql.DB
}

// NewDeviceRepository builds a Postgres-backed device repository.
func NewDeviceRepository(db *sql.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

const deviceColumns = `id, user_id, device_id, name, refresh_token_id, created_at, last_seen_at`

func scanDevice(scanner interface {
	Scan(dest ...any) error
}) (*user.Device, error) {
	var d user.Device
	var refreshID string
	err := scanner.Scan(
		&d.ID, &d.UserID, &d.DeviceID, &d.Name, &refreshID, &d.CreatedAt, &d.LastSeenAt,
	)
	if err != nil {
		return nil, err
	}
	d.RefreshTokenID = user.RefreshTokenID(refreshID)
	return &d, nil
}

// FindByUserAndDevice returns the device row for a user+client device id.
// Missing rows return (nil, nil).
func (r *DeviceRepository) FindByUserAndDevice(ctx context.Context, userID user.ID, deviceID user.DeviceID) (*user.Device, error) {
	const q = `SELECT ` + deviceColumns + ` FROM user_devices WHERE user_id = $1 AND device_id = $2`
	d, err := scanDevice(r.db.QueryRowContext(ctx, q, userID, deviceID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to look up device", err)
	}
	return d, nil
}

// FindByID returns a device by its server row id.
func (r *DeviceRepository) FindByID(ctx context.Context, id user.DeviceRowID) (*user.Device, error) {
	const q = `SELECT ` + deviceColumns + ` FROM user_devices WHERE id = $1`
	d, err := scanDevice(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.CodeNotFound, "device not found")
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to look up device", err)
	}
	return d, nil
}

// ListByUser returns devices newest-last-seen first.
func (r *DeviceRepository) ListByUser(ctx context.Context, userID user.ID) ([]*user.Device, error) {
	const q = `SELECT ` + deviceColumns + ` FROM user_devices WHERE user_id = $1 ORDER BY last_seen_at DESC`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list devices", err)
	}
	defer rows.Close()

	var out []*user.Device
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan device", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to iterate devices", err)
	}
	if out == nil {
		out = []*user.Device{}
	}
	return out, nil
}

// CountByUser returns how many devices the user currently has registered.
func (r *DeviceRepository) CountByUser(ctx context.Context, userID user.ID) (int, error) {
	const q = `SELECT COUNT(*) FROM user_devices WHERE user_id = $1`
	var n int
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&n); err != nil {
		return 0, apperrors.Wrap(apperrors.CodeInternal, "failed to count devices", err)
	}
	return n, nil
}

// Create inserts a new device row.
func (r *DeviceRepository) Create(ctx context.Context, d *user.Device) error {
	const q = `
		INSERT INTO user_devices (id, user_id, device_id, name, refresh_token_id, created_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, q,
		d.ID, d.UserID, d.DeviceID, d.Name, string(d.RefreshTokenID), d.CreatedAt, d.LastSeenAt,
	)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create device", err)
	}
	return nil
}

// UpdateSession refreshes name, refresh token id, and last_seen_at.
func (r *DeviceRepository) UpdateSession(ctx context.Context, d *user.Device) error {
	const q = `
		UPDATE user_devices
		SET name = $2, refresh_token_id = $3, last_seen_at = $4
		WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, d.ID, d.Name, string(d.RefreshTokenID), d.LastSeenAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update device", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm device update", err)
	}
	if n == 0 {
		return apperrors.New(apperrors.CodeNotFound, "device not found")
	}
	return nil
}

// Delete removes a device row by server id.
func (r *DeviceRepository) Delete(ctx context.Context, id user.DeviceRowID) error {
	const q = `DELETE FROM user_devices WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to delete device", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm device delete", err)
	}
	if n == 0 {
		return apperrors.New(apperrors.CodeNotFound, "device not found")
	}
	return nil
}
