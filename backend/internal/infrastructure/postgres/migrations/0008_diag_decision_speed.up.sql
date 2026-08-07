-- TASK-01: decision rule/reason + transfer speed on diag_events.
ALTER TABLE diag_events ADD COLUMN IF NOT EXISTS rule TEXT NOT NULL DEFAULT '';
ALTER TABLE diag_events ADD COLUMN IF NOT EXISTS decision_reason TEXT NOT NULL DEFAULT '';
ALTER TABLE diag_events ADD COLUMN IF NOT EXISTS speed_kbps INTEGER NOT NULL DEFAULT 0;
