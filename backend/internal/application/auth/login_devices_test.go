package auth_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"streampass/backend/internal/application/auth"
	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

type memDevices struct {
	mu   sync.Mutex
	byID map[user.DeviceRowID]*user.Device
}

func newMemDevices() *memDevices {
	return &memDevices{byID: map[user.DeviceRowID]*user.Device{}}
}

func (m *memDevices) FindByUserAndDevice(_ context.Context, userID user.ID, deviceID user.DeviceID) (*user.Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.byID {
		if d.UserID == userID && d.DeviceID == deviceID {
			cp := *d
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memDevices) FindByID(_ context.Context, id user.DeviceRowID) (*user.Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.byID[id]
	if !ok {
		return nil, apperrors.New(apperrors.CodeNotFound, "device not found")
	}
	cp := *d
	return &cp, nil
}

func (m *memDevices) ListByUser(_ context.Context, userID user.ID) ([]*user.Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*user.Device
	for _, d := range m.byID {
		if d.UserID == userID {
			cp := *d
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *memDevices) CountByUser(_ context.Context, userID user.ID) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, d := range m.byID {
		if d.UserID == userID {
			n++
		}
	}
	return n, nil
}

func (m *memDevices) Create(_ context.Context, d *user.Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *d
	m.byID[d.ID] = &cp
	return nil
}

func (m *memDevices) UpdateSession(_ context.Context, d *user.Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.byID[d.ID]
	if !ok {
		return apperrors.New(apperrors.CodeNotFound, "device not found")
	}
	cur.Name = d.Name
	cur.RefreshTokenID = d.RefreshTokenID
	cur.LastSeenAt = d.LastSeenAt
	return nil
}

func (m *memDevices) Delete(_ context.Context, id user.DeviceRowID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[id]; !ok {
		return apperrors.New(apperrors.CodeNotFound, "device not found")
	}
	delete(m.byID, id)
	return nil
}

type stubTokens struct {
	n int
}

func (s *stubTokens) IssueAccessToken(user.ID) (string, time.Time, error) {
	s.n++
	return "access", time.Now().UTC().Add(time.Hour), nil
}
func (s *stubTokens) IssueRefreshToken(user.ID) (string, user.RefreshTokenID, time.Time, error) {
	s.n++
	id := user.RefreshTokenID("rt-" + time.Now().Format("150405.000000000"))
	return "refresh", id, time.Now().UTC().Add(24 * time.Hour), nil
}
func (s *stubTokens) ParseAccessToken(string) (user.ID, error) { return "", nil }
func (s *stubTokens) ParseRefreshToken(string) (user.ID, user.RefreshTokenID, error) {
	return "", "", nil
}

func TestLoginDeviceLimit(t *testing.T) {
	users := newMemUsers()
	devices := newMemDevices()
	now := time.Now().UTC()
	u := user.NewUser("u1", "a@b.c", "h:pass", now)
	_ = users.Create(context.Background(), u)

	uc := auth.NewLoginUseCase(
		users, devices, plainHasher{}, &stubTokens{}, &memSessions{},
		auth.SystemClock{}, 2, logger.New("test", "error"),
	)

	for i, id := range []string{"d1", "d2"} {
		_, err := uc.Execute(context.Background(), auth.LoginInput{
			Email: "a@b.c", Password: "pass", DeviceID: id, DeviceName: "Phone",
		})
		if err != nil {
			t.Fatalf("login device %d: %v", i+1, err)
		}
	}

	_, err := uc.Execute(context.Background(), auth.LoginInput{
		Email: "a@b.c", Password: "pass", DeviceID: "d3", DeviceName: "Tablet",
	})
	if err == nil {
		t.Fatal("expected device limit error")
	}
	if apperrors.CodeOf(err) != apperrors.CodeDeviceLimit {
		t.Fatalf("code = %s, want AUTH_DEVICE_LIMIT", apperrors.CodeOf(err))
	}

	// Same device re-login must not hit the limit.
	_, err = uc.Execute(context.Background(), auth.LoginInput{
		Email: "a@b.c", Password: "pass", DeviceID: "d1", DeviceName: "Phone",
	})
	if err != nil {
		t.Fatalf("re-login same device: %v", err)
	}

	listUC := auth.NewListDevicesUseCase(devices, 2, logger.New("test", "error"))
	list, err := listUC.Execute(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Devices) != 2 {
		t.Fatalf("devices = %d, want 2", len(list.Devices))
	}

	revokeUC := auth.NewRevokeDeviceUseCase(devices, &memSessions{}, logger.New("test", "error"))
	if err := revokeUC.Execute(context.Background(), u.ID, list.Devices[0].ID); err != nil {
		t.Fatal(err)
	}

	_, err = uc.Execute(context.Background(), auth.LoginInput{
		Email: "a@b.c", Password: "pass", DeviceID: "d3", DeviceName: "Tablet",
	})
	if err != nil {
		t.Fatalf("login after revoke: %v", err)
	}
}
