-- Enrich operator diagnostics with readable site label, reason, slow flag.
ALTER TABLE diag_events ADD COLUMN IF NOT EXISTS site TEXT NOT NULL DEFAULT '';
ALTER TABLE diag_events ADD COLUMN IF NOT EXISTS reason TEXT NOT NULL DEFAULT '';
ALTER TABLE diag_events ADD COLUMN IF NOT EXISTS slow BOOLEAN NOT NULL DEFAULT false;
