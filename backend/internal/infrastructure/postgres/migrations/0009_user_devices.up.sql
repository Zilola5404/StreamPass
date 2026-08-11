-- Per-user registered devices for multi-device limit (BL-049 / E10).
CREATE TABLE IF NOT EXISTS user_devices (
    id                TEXT PRIMARY KEY,
    user_id           TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    device_id         TEXT NOT NULL,
    name              TEXT NOT NULL DEFAULT '',
    refresh_token_id  TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL,
    last_seen_at      TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices (user_id);
