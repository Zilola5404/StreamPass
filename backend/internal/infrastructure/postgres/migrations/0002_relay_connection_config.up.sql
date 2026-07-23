-- Adds the opaque connection config a client needs to actually connect to
-- a relay (e.g. a Hiddify/sing-box subscription URL). Defaults to empty
-- string so existing rows (registered before this column existed) don't
-- break — an empty config just means "not yet provisioned for clients".
ALTER TABLE relay_servers
    ADD COLUMN IF NOT EXISTS connection_config TEXT NOT NULL DEFAULT '';
