// Package configsvc is the Config Service application layer: serving the
// current client configuration and letting an operator publish a new
// version (spec: "Config Service" module, mirrors the Rule Service's
// publish/version pattern).
package configsvc

import (
	"context"
	"errors"
	"time"

	"streampass/backend/internal/domain/appconfig"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

// Clock is injected for deterministic tests.
type Clock interface {
	Now() time.Time
}

// SystemClock is the production Clock implementation.
type SystemClock struct{}

// Now returns the current wall-clock time.
func (SystemClock) Now() time.Time { return time.Now().UTC() }

// Service implements the Config Service use cases.
type Service struct {
	repo  appconfig.Repository
	clock Clock
	log   *logger.Logger
}

// NewService wires the Config Service via constructor injection.
func NewService(repo appconfig.Repository, clock Clock, log *logger.Logger) *Service {
	return &Service{repo: repo, clock: clock, log: log.With("config_service")}
}

// GetLatest implements "GET /config". If no config has ever been
// published, this is a legitimate, expected state (not a server failure)
// — the caller sees a 404 with a clear message rather than an opaque 500.
func (s *Service) GetLatest(ctx context.Context) (*appconfig.Config, error) {
	cfg, err := s.repo.Latest(ctx)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == apperrors.CodeNotFound {
			return nil, err
		}
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to load config", err)
	}
	return cfg, nil
}

// Publish validates and stores a new configuration version. Used by the
// Admin Panel / operator tooling.
func (s *Service) Publish(ctx context.Context, cfg appconfig.Config) (*appconfig.Config, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	published, err := s.repo.Publish(ctx, cfg, s.clock.Now())
	if err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to publish config", err)
	}
	return published, nil
}

func validateConfig(cfg appconfig.Config) error {
	if cfg.MinSupportedClientVer == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "min_supported_client_version must not be empty").
			WithDetails(map[string]any{"field": "min_supported_client_version"})
	}
	if cfg.RulePollIntervalSec <= 0 {
		return apperrors.New(apperrors.CodeInvalidInput, "rule_poll_interval_sec must be positive").
			WithDetails(map[string]any{"field": "rule_poll_interval_sec"})
	}
	if cfg.RelayPollIntervalSec <= 0 {
		return apperrors.New(apperrors.CodeInvalidInput, "relay_poll_interval_sec must be positive").
			WithDetails(map[string]any{"field": "relay_poll_interval_sec"})
	}
	return nil
}
