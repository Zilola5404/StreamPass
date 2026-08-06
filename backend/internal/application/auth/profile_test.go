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

type memUsers struct {
	mu    sync.Mutex
	byID  map[user.ID]*user.User
	byMail map[string]*user.User
}

func newMemUsers() *memUsers {
	return &memUsers{byID: map[user.ID]*user.User{}, byMail: map[string]*user.User{}}
}

func (m *memUsers) Create(_ context.Context, u *user.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *u
	m.byID[u.ID] = &cp
	m.byMail[u.Email] = &cp
	return nil
}
func (m *memUsers) FindByEmail(_ context.Context, email string) (*user.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byMail[email]
	if !ok {
		return nil, user.ErrNotFound(email)
	}
	cp := *u
	return &cp, nil
}
func (m *memUsers) FindByID(_ context.Context, id user.ID) (*user.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return nil, user.ErrNotFound(string(id))
	}
	cp := *u
	return &cp, nil
}
func (m *memUsers) ExtendSubscription(context.Context, user.ID, time.Time) error { return nil }
func (m *memUsers) List(context.Context) ([]*user.User, error)                    { return nil, nil }
func (m *memUsers) UpdatePasswordHash(_ context.Context, id user.ID, hash string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return user.ErrNotFound(string(id))
	}
	u.PasswordHash = hash
	u.UpdatedAt = now
	return nil
}
func (m *memUsers) Delete(_ context.Context, id user.ID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return user.ErrNotFound(string(id))
	}
	delete(m.byMail, u.Email)
	delete(m.byID, id)
	return nil
}

type plainHasher struct{}

func (plainHasher) Hash(p string) (string, error) { return "h:" + p, nil }
func (plainHasher) Verify(hash, p string) bool     { return hash == "h:"+p }

type memSessions struct {
	revoked int
}

func (m *memSessions) Store(context.Context, user.ID, user.RefreshTokenID, time.Duration) error {
	return nil
}
func (m *memSessions) IsValid(context.Context, user.ID, user.RefreshTokenID) (bool, error) {
	return true, nil
}
func (m *memSessions) Revoke(context.Context, user.ID, user.RefreshTokenID) error { return nil }
func (m *memSessions) RevokeAll(context.Context, user.ID) error {
	m.revoked++
	return nil
}

type memReset struct {
	mu   sync.Mutex
	data map[string]user.ID
}

func newMemReset() *memReset { return &memReset{data: map[string]user.ID{}} }
func (m *memReset) Save(_ context.Context, token string, id user.ID, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[token] = id
	return nil
}
func (m *memReset) Consume(_ context.Context, token string) (user.ID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.data[token]
	if !ok {
		return "", apperrors.New(apperrors.CodeNotFound, "missing")
	}
	delete(m.data, token)
	return id, nil
}

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func TestForgotAndResetPassword(t *testing.T) {
	log := logger.New("test", "error")
	users := newMemUsers()
	u := user.NewUser("u1", "a@b.co", "h:oldpass12", time.Now().UTC())
	_ = users.Create(context.Background(), u)
	tokens := newMemReset()
	sessions := &memSessions{}
	clock := fixedClock{t: time.Now().UTC()}

	forgot := auth.NewForgotPasswordUseCase(users, tokens, true, log)
	res, err := forgot.Execute(context.Background(), "a@b.co")
	if err != nil {
		t.Fatal(err)
	}
	if res.ResetToken == "" {
		t.Fatal("expected exposed reset token")
	}

	reset := auth.NewResetPasswordUseCase(users, plainHasher{}, tokens, sessions, clock, log)
	if err := reset.Execute(context.Background(), res.ResetToken, "newpass99"); err != nil {
		t.Fatal(err)
	}
	got, _ := users.FindByID(context.Background(), "u1")
	if got.PasswordHash != "h:newpass99" {
		t.Fatalf("hash=%s", got.PasswordHash)
	}
	if sessions.revoked != 1 {
		t.Fatalf("revokeAll calls=%d", sessions.revoked)
	}
}

func TestChangePasswordAndDelete(t *testing.T) {
	log := logger.New("test", "error")
	users := newMemUsers()
	u := user.NewUser("u1", "a@b.co", "h:oldpass12", time.Now().UTC())
	_ = users.Create(context.Background(), u)
	sessions := &memSessions{}
	clock := fixedClock{t: time.Now().UTC()}

	chg := auth.NewChangePasswordUseCase(users, plainHasher{}, sessions, clock, log)
	if err := chg.Execute(context.Background(), "u1", "oldpass12", "newerpass"); err != nil {
		t.Fatal(err)
	}
	del := auth.NewDeleteAccountUseCase(users, sessions, log)
	if err := del.Execute(context.Background(), "u1"); err != nil {
		t.Fatal(err)
	}
	if _, err := users.FindByID(context.Background(), "u1"); err == nil {
		t.Fatal("expected deleted")
	}
}
