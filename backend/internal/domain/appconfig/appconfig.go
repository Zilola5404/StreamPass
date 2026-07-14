// Package appconfig contains the Config Service's domain model: dynamic,
// versioned client configuration (spec: "Config Service" module) — distinct
// from the Rule Service (routing rules) and Relay Manager (server list).
// Config Service governs client-wide behavioral knobs the operator wants
// to change without shipping a new client build: minimum supported client
// version, telemetry on/off, poll intervals.
package appconfig

import (
	"context"
	"time"

	apperrors "streampass/shared/errors"
)

// Config is a single versioned snapshot of client-facing configuration.
type Config struct {
	Version               int
	MinSupportedClientVer string
	TelemetryEnabled      bool
	RulePollIntervalSec   int
	RelayPollIntervalSec  int
	UpdatedAt             time.Time
}

// Repository is the port the Config Service application layer depends on.
type Repository interface {
	Latest(ctx context.Context) (*Config, error)
	Publish(ctx context.Context, c Config, publishedAt time.Time) (*Config, error)
}

// ErrNoConfig is returned when no configuration has been published yet.
var ErrNoConfig = apperrors.New(apperrors.CodeNotFound, "no config has been published yet")
