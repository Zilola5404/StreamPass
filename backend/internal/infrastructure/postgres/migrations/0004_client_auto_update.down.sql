ALTER TABLE app_configs
    DROP COLUMN IF EXISTS latest_client_version,
    DROP COLUMN IF EXISTS client_download_url;
