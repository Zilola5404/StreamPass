-- Operator diagnostic events (hostname/IP/mode/latency). Not full URLs or page history.
-- Retention is enforced by application purge (7 days). Admin-read only.
CREATE TABLE IF NOT EXISTS diag_events (
    id              BIGSERIAL PRIMARY KEY,
    user_id         TEXT NOT NULL,
    proto           TEXT NOT NULL DEFAULT 'tcp',
    host            TEXT NOT NULL DEFAULT '',
    dest_ip         TEXT NOT NULL DEFAULT '',
    dest_port       INTEGER NOT NULL DEFAULT 0,
    mode            TEXT NOT NULL DEFAULT '',
    result          TEXT NOT NULL DEFAULT '',
    latency_ms      INTEGER NOT NULL DEFAULT 0,
    error_code      TEXT NOT NULL DEFAULT '',
    relay_id        TEXT NOT NULL DEFAULT '',
    client_version  TEXT NOT NULL DEFAULT '',
    recorded_at     TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_diag_events_recorded_at ON diag_events (recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_diag_events_user_recorded ON diag_events (user_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_diag_events_result ON diag_events (result, recorded_at DESC);
