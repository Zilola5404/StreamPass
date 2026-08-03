// Command server is the StreamPass backend's composition root: it loads
// config, constructs every infrastructure adapter and application service,
// wires them together via constructor injection, and starts the HTTP
// server. No business logic lives here — only wiring (Clean Architecture:
// the one place allowed to know about every layer at once).
package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authsvc "streampass/backend/internal/application/auth"
	adminsvc "streampass/backend/internal/application/admin"
	billingsvc "streampass/backend/internal/application/billing"
	configsvcpkg "streampass/backend/internal/application/configsvc"
	relaysvc "streampass/backend/internal/application/relay"
	rulesvc "streampass/backend/internal/application/rule"
	telemetrysvc "streampass/backend/internal/application/telemetry"
	"streampass/backend/internal/domain/user"
	"streampass/backend/internal/infrastructure/http/handler"
	"streampass/backend/internal/infrastructure/http/router"
	"streampass/backend/internal/infrastructure/payment/yookassa"
	"streampass/backend/internal/infrastructure/postgres"
	"streampass/backend/internal/infrastructure/redisclient"
	"streampass/backend/internal/infrastructure/security"
	"streampass/shared/config"
	"streampass/shared/idgen"
	"streampass/shared/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	configPath := envOr("CONFIG_PATH", "config.yaml")
	cfg, err := config.NewFileLoader(configPath).Load()
	if err != nil {
		return err
	}

	log := logger.New("backend", cfg.StringOr("server.log_level", "info"))
	ctx := context.Background()

	db, err := postgres.Open(postgres.PoolConfig{
		DSN:             buildPostgresDSN(cfg),
		MaxOpenConns:    cfg.IntOr("database.max_open_conns", 20),
		MaxIdleConns:    cfg.IntOr("database.max_idle_conns", 5),
		ConnMaxLifetime: cfg.DurationOr("database.conn_max_lifetime", 30*time.Minute),
	})
	if err != nil {
		return err
	}
	defer db.Close()

	if err := postgres.Migrate(ctx, db); err != nil {
		return err
	}

	redis := redisclient.New(redisclient.Config{
		Addr:     cfg.StringOr("redis.addr", "localhost:6379"),
		Password: cfg.StringOr("redis.password", ""),
	})

	deps := buildDeps(cfg, db, redis, log)

	srv := &http.Server{
		Addr:              ":" + intToStr(cfg.IntOr("server.http_port", 8080)),
		Handler:           router.New(deps),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return serveWithGracefulShutdown(ctx, srv, log)
}

// buildDeps constructs every infrastructure adapter and application
// service and returns the router.Deps struct main() hands to router.New.
// Kept as one function (rather than spread across main) so the full
// dependency graph is visible in one place.
func buildDeps(cfg *config.Config, db *sql.DB, redis *redisclient.Client, log *logger.Logger) router.Deps {
	// --- Infrastructure adapters ---
	userRepo := postgres.NewUserRepository(db)
	ruleRepo := postgres.NewRuleRepository(db)
	relayRepo := postgres.NewRelayRepository(db)
	telemetryRepo := postgres.NewTelemetryRepository(db)
	appConfigRepo := postgres.NewAppConfigRepository(db)
	paymentRepo := postgres.NewPaymentRepository(db)

	hasher := security.NewArgon2Hasher()
	tokens := security.NewJWTTokenIssuer(
		cfg.StringOr("jwt.secret", ""),
		cfg.DurationOr("jwt.access_ttl", 15*time.Minute),
		cfg.DurationOr("jwt.refresh_ttl", 720*time.Hour),
	)
	sessions := redisclient.NewSessionStore(redis)

	paymentProvider := yookassa.New(yookassa.Config{
		ShopID:    cfg.StringOr("billing.yookassa_shop_id", ""),
		SecretKey: cfg.StringOr("billing.yookassa_secret_key", ""),
		ReturnURL: cfg.StringOr("billing.yookassa_return_url", ""),
	})

	// --- Application services ---
	registerUC := authsvc.NewRegisterUseCase(userRepo, hasher, idGeneratorAdapter{}, authsvc.SystemClock{}, log)
	loginUC := authsvc.NewLoginUseCase(userRepo, hasher, tokens, sessions, authsvc.SystemClock{}, log)
	logoutUC := authsvc.NewLogoutUseCase(tokens, sessions, log)
	refreshUC := authsvc.NewRefreshUseCase(tokens, sessions, log)
	authService := authsvc.NewService(registerUC, loginUC, logoutUC, refreshUC)

	ruleService := rulesvc.NewService(ruleRepo, rulesvc.SystemClock{}, log)
	relayService := relaysvc.NewService(relayRepo, log)
	telemetryService := telemetrysvc.NewService(telemetryRepo, telemetrysvc.SystemClock{}, log)
	configService := configsvcpkg.NewService(appConfigRepo, configsvcpkg.SystemClock{}, log)
	adminUserService := adminsvc.NewUserService(userRepo, adminsvc.SystemClock{}, log)
	billingService := billingsvc.NewService(userRepo, paymentRepo, paymentProvider, billingsvc.Plan{
		AmountRUB:  int64(cfg.IntOr("billing.plan_amount_rub", 299)),
		PeriodDays: cfg.IntOr("billing.plan_period_days", 30),
	}, billingsvc.SystemClock{}, log)

	return router.Deps{
		Auth:            handler.NewAuthHandler(authService),
		Rule:            handler.NewRuleHandler(ruleService),
		Relay:           handler.NewRelayHandler(relayService),
		Telemetry:       handler.NewTelemetryHandler(telemetryService),
		Config:          handler.NewConfigHandler(configService),
		Billing:         handler.NewBillingHandler(billingService),
		Health:          handler.NewHealthHandler(),
		Admin:           handler.NewAdminHandler(adminUserService),
		TokenVerifier:   tokens,
		AdminKey:        cfg.StringOr("admin.api_key", ""),
		Log:             log,
		PublicRateLimit: cfg.IntOr("rate_limit.public_requests_per_window", 20),
		PublicWindow:    cfg.DurationOr("rate_limit.public_window", time.Minute),
	}
}

// idGeneratorAdapter adapts shared/idgen to the auth.IDGenerator port.
type idGeneratorAdapter struct{}

func (idGeneratorAdapter) NewID() user.ID { return user.ID(idgen.New()) }

func buildPostgresDSN(cfg *config.Config) string {
	return "host=" + cfg.StringOr("database.host", "localhost") +
		" port=" + intToStr(cfg.IntOr("database.port", 5432)) +
		" dbname=" + cfg.StringOr("database.name", "streampass") +
		" user=" + cfg.StringOr("database.user", "streampass") +
		" password=" + cfg.StringOr("database.password", "") +
		" sslmode=" + cfg.StringOr("database.sslmode", "disable")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// serveWithGracefulShutdown runs srv until SIGINT/SIGTERM, then drains
// in-flight requests within a bounded timeout before returning.
func serveWithGracefulShutdown(ctx context.Context, srv *http.Server, log *logger.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		log.Info(ctx, "http server starting", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		log.Info(ctx, "shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
