// Package integrationtest holds HTTP+Postgres integration tests (BL-011).
// Tests skip cleanly when Docker is unavailable or -short is set.
package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	adminsvc "streampass/backend/internal/application/admin"
	authsvc "streampass/backend/internal/application/auth"
	billingsvc "streampass/backend/internal/application/billing"
	configsvcpkg "streampass/backend/internal/application/configsvc"
	exclusionsvc "streampass/backend/internal/application/exclusion"
	relaysvc "streampass/backend/internal/application/relay"
	rulesvc "streampass/backend/internal/application/rule"
	telemetrysvc "streampass/backend/internal/application/telemetry"
	"streampass/backend/internal/domain/user"
	"streampass/backend/internal/infrastructure/http/handler"
	"streampass/backend/internal/infrastructure/http/router"
	"streampass/backend/internal/infrastructure/postgres"
	"streampass/backend/internal/infrastructure/security"
	"streampass/shared/idgen"
	"streampass/shared/logger"
	apperrors "streampass/shared/errors"
)

const (
	testAdminKey  = "test-admin-key"
	testJWTSecret = "integration-test-jwt-secret-32b!!"
)

// RequireDocker skips the test when Docker is not usable.
func RequireDocker(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skipf("docker not available: %v", err)
	}
}

// StartPostgres starts postgres:16-alpine and returns a migrated *sql.DB.
func StartPostgres(t *testing.T) *sql.DB {
	t.Helper()
	RequireDocker(t)
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "streampass",
			"POSTGRES_PASSWORD": "streampass",
			"POSTGRES_DB":       "streampass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("postgres host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("postgres port: %v", err)
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=streampass password=streampass dbname=streampass sslmode=disable",
		host, port.Port(),
	)

	db, err := postgres.Open(postgres.PoolConfig{
		DSN:          dsn,
		MaxOpenConns: 5,
		MaxIdleConns: 2,
		ConnMaxLifetime: time.Minute,
	})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := postgres.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// memorySessions is an in-memory user.SessionStore for tests (no Redis).
type memorySessions struct {
	mu   sync.Mutex
	data map[string]time.Time // key: userID|tokenID → expiry
}

func newMemorySessions() *memorySessions {
	return &memorySessions{data: make(map[string]time.Time)}
}

func sessionKey(userID user.ID, tokenID user.RefreshTokenID) string {
	return string(userID) + "|" + string(tokenID)
}

func (s *memorySessions) Store(_ context.Context, userID user.ID, tokenID user.RefreshTokenID, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[sessionKey(userID, tokenID)] = time.Now().UTC().Add(ttl)
	return nil
}

func (s *memorySessions) IsValid(_ context.Context, userID user.ID, tokenID user.RefreshTokenID) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.data[sessionKey(userID, tokenID)]
	if !ok {
		return false, nil
	}
	if time.Now().UTC().After(exp) {
		delete(s.data, sessionKey(userID, tokenID))
		return false, nil
	}
	return true, nil
}

func (s *memorySessions) Revoke(_ context.Context, userID user.ID, tokenID user.RefreshTokenID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, sessionKey(userID, tokenID))
	return nil
}

func (s *memorySessions) RevokeAll(_ context.Context, userID user.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := string(userID) + "|"
	for k := range s.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(s.data, k)
		}
	}
	return nil
}

// memoryResetTokens is an in-memory auth.ResetTokenStore for tests.
type memoryResetTokens struct {
	mu   sync.Mutex
	data map[string]user.ID
}

func newMemoryResetTokens() *memoryResetTokens {
	return &memoryResetTokens{data: make(map[string]user.ID)}
}

func (s *memoryResetTokens) Save(_ context.Context, token string, userID user.ID, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[token] = userID
	return nil
}

func (s *memoryResetTokens) Consume(_ context.Context, token string) (user.ID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.data[token]
	if !ok {
		return "", apperrors.New(apperrors.CodeNotFound, "reset token not found")
	}
	delete(s.data, token)
	return id, nil
}

type idGen struct{}

func (idGen) NewID() user.ID { return user.ID(idgen.New()) }

type fakePayments struct {
	mu       sync.Mutex
	payments map[string]billingsvc.PaymentStatus
}

func newFakePayments() *fakePayments {
	return &fakePayments{payments: make(map[string]billingsvc.PaymentStatus)}
}

func (f *fakePayments) CreatePayment(_ context.Context, userID string, amountRUB int64, description string) (string, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := "pay_" + idgen.New()[:16]
	f.payments[id] = billingsvc.PaymentStatusPending
	return id, "https://example.test/pay/" + id, nil
}

func (f *fakePayments) FetchPaymentStatus(_ context.Context, providerPaymentID string) (billingsvc.PaymentStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	st, ok := f.payments[providerPaymentID]
	if !ok {
		return "", fmt.Errorf("unknown payment")
	}
	return st, nil
}

func (f *fakePayments) Succeed(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.payments[id] = billingsvc.PaymentStatusSucceeded
}

// NewTestHandler wires a full HTTP API against the given DB (mirrors buildDeps).
func NewTestHandler(t *testing.T, db *sql.DB) (http.Handler, *fakePayments) {
	t.Helper()
	log := logger.New("integrationtest", "error")
	sessions := newMemorySessions()
	payments := newFakePayments()

	userRepo := postgres.NewUserRepository(db)
	ruleRepo := postgres.NewRuleRepository(db)
	relayRepo := postgres.NewRelayRepository(db)
	telemetryRepo := postgres.NewTelemetryRepository(db)
	appConfigRepo := postgres.NewAppConfigRepository(db)
	paymentRepo := postgres.NewPaymentRepository(db)
	exclusionRepo := postgres.NewExclusionRepository(db)

	hasher := security.NewArgon2Hasher()
	tokens := security.NewJWTTokenIssuer(testJWTSecret, 15*time.Minute, 720*time.Hour)

	registerUC := authsvc.NewRegisterUseCase(userRepo, hasher, idGen{}, authsvc.SystemClock{}, log)
	loginUC := authsvc.NewLoginUseCase(userRepo, hasher, tokens, sessions, authsvc.SystemClock{}, log)
	logoutUC := authsvc.NewLogoutUseCase(tokens, sessions, log)
	refreshUC := authsvc.NewRefreshUseCase(tokens, sessions, log)
	resetTokens := newMemoryResetTokens()
	authService := authsvc.NewService(
		registerUC, loginUC, logoutUC, refreshUC,
		authsvc.NewGetProfileUseCase(userRepo, log),
		authsvc.NewChangePasswordUseCase(userRepo, hasher, sessions, authsvc.SystemClock{}, log),
		authsvc.NewDeleteAccountUseCase(userRepo, sessions, log),
		authsvc.NewForgotPasswordUseCase(userRepo, resetTokens, true, log),
		authsvc.NewResetPasswordUseCase(userRepo, hasher, resetTokens, sessions, authsvc.SystemClock{}, log),
	)

	billingService := billingsvc.NewService(userRepo, paymentRepo, payments, []billingsvc.Plan{
		{Code: "month", Title: "Месяц", AmountRUB: 299, PeriodDays: 30},
		{Code: "year", Title: "Год", AmountRUB: 2990, PeriodDays: 365},
	}, billingsvc.SystemClock{}, log)

	h := router.New(router.Deps{
		Auth:            handler.NewAuthHandler(authService),
		Rule:            handler.NewRuleHandler(rulesvc.NewService(ruleRepo, rulesvc.SystemClock{}, log)),
		Relay:           handler.NewRelayHandler(relaysvc.NewService(relayRepo, log)),
		Telemetry:       handler.NewTelemetryHandler(telemetrysvc.NewService(telemetryRepo, telemetrysvc.SystemClock{}, log)),
		Config:          handler.NewConfigHandler(configsvcpkg.NewService(appConfigRepo, configsvcpkg.SystemClock{}, log)),
		Billing:         handler.NewBillingHandler(billingService, ""),
		Exclusion:       handler.NewExclusionHandler(exclusionsvc.NewService(exclusionRepo, log)),
		Health:          handler.NewHealthHandler(),
		Admin:           handler.NewAdminHandler(adminsvc.NewUserService(userRepo, adminsvc.SystemClock{}, log)),
		TokenVerifier:   tokens,
		AdminKey:        testAdminKey,
		Log:             log,
		PublicRateLimit: 1000,
		PublicWindow:    time.Minute,
	})
	return h, payments
}
