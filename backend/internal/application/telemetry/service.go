// Package telemetry is the Telemetry application layer: ingesting
// technical events from clients (spec section 14, "POST /telemetry").
package telemetry

import (
	"context"
	"time"

	"streampass/backend/internal/domain/telemetry"
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

// Service implements the Telemetry use cases.
type Service struct {
	repo  telemetry.Repository
	clock Clock
	log   *logger.Logger
}

// NewService wires the Telemetry service via constructor injection.
func NewService(repo telemetry.Repository, clock Clock, log *logger.Logger) *Service {
	return &Service{repo: repo, clock: clock, log: log.With("telemetry_service")}
}

// Record validates and persists a technical telemetry event. Rejects
// anything that isn't a plain technical value, per the spec's "no PII"
// requirement — the type system already prevents callers from attaching a
// URL or browsing history, since Event has no field for it.
func (s *Service) Record(ctx context.Context, e telemetry.Event) error {
	if e.UserID == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "telemetry event missing user id")
	}
	if e.RTTMillis < 0 || e.ConnectMillis < 0 {
		return apperrors.New(apperrors.CodeInvalidInput, "telemetry event has negative timing value")
	}
	if e.PacketLossPct < 0 || e.PacketLossPct > 100 {
		return apperrors.New(apperrors.CodeInvalidInput, "telemetry event packet loss out of range")
	}

	e.RecordedAt = s.clock.Now()
	if err := s.repo.Record(ctx, e); err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to record telemetry event", err)
	}
	return nil
}
