-- BL-022: client auto-update metadata on published configs.
ALTER TABLE app_configs
    ADD COLUMN IF NOT EXISTS latest_client_version TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS client_download_url TEXT NOT NULL DEFAULT '';
