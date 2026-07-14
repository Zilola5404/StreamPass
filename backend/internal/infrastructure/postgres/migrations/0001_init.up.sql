-- StreamPass backend: initial schema.
-- One migration per bounded context's storage needs. Applied in order by
-- filename; see backend/migrations/README.md for how to run these.

CREATE TABLE IF NOT EXISTS users (
    id                          TEXT PRIMARY KEY,
    email                       TEXT NOT NULL UNIQUE,
    password_hash               TEXT NOT NULL,
    created_at                  TIMESTAMPTZ NOT NULL,
    updated_at                  TIMESTAMPTZ NOT NULL,
    subscription_active_until   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE TABLE IF NOT EXISTS payments (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users (id),
    provider_id  TEXT NOT NULL UNIQUE,
    amount_rub   BIGINT NOT NULL,
    period_days  INTEGER NOT NULL,
    status       TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments (user_id);
CREATE INDEX IF NOT EXISTS idx_payments_provider_id ON payments (provider_id);

-- One immutable row per published rule-set version (spec section 6: Rule
-- Service). Rules are stored as a JSONB array rather than normalized rows
-- since the whole set must be atomic per version and MVP scale is a few
-- hundred rules at most (KISS).
CREATE TABLE IF NOT EXISTS rule_sets (
    version     INTEGER PRIMARY KEY,
    rules       JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);

-- One immutable row per published client-config version (Config Service),
-- mirroring rule_sets' versioning pattern.
CREATE TABLE IF NOT EXISTS app_configs (
    version                          INTEGER PRIMARY KEY,
    min_supported_client_version     TEXT NOT NULL,
    telemetry_enabled                BOOLEAN NOT NULL,
    rule_poll_interval_sec           INTEGER NOT NULL,
    relay_poll_interval_sec          INTEGER NOT NULL,
    updated_at                       TIMESTAMPTZ NOT NULL
);

-- Relay Manager's server registry (spec section 9), continuously updated
-- by the Health Monitor.
CREATE TABLE IF NOT EXISTS relay_servers (
    id          TEXT PRIMARY KEY,
    region      TEXT NOT NULL,
    host        TEXT NOT NULL,
    port        INTEGER NOT NULL,
    healthy     BOOLEAN NOT NULL DEFAULT false,
    load_ratio  DOUBLE PRECISION NOT NULL DEFAULT 0,
    rtt_millis  INTEGER NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL
);

-- Technical telemetry events (spec section 14). No PII, no URLs, no
-- browsing history columns exist here by design — see
-- backend/internal/domain/telemetry/event.go.
CREATE TABLE IF NOT EXISTS telemetry_events (
    id              BIGSERIAL PRIMARY KEY,
    user_id         TEXT NOT NULL,
    rtt_millis      INTEGER NOT NULL,
    packet_loss_pct DOUBLE PRECISION NOT NULL,
    relay_id        TEXT NOT NULL,
    client_version  TEXT NOT NULL,
    os              TEXT NOT NULL,
    connect_millis  INTEGER NOT NULL,
    error_code      TEXT NOT NULL DEFAULT '',
    recorded_at     TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_telemetry_events_user_id ON telemetry_events (user_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_events_recorded_at ON telemetry_events (recorded_at);
