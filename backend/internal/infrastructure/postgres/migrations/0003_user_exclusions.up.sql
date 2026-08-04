-- User-level DIRECT domain exclusions (BL-014), synced from the client.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS exclusions JSONB NOT NULL DEFAULT '[]';
