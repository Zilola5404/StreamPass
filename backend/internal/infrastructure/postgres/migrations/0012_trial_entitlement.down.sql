ALTER TABLE users
    DROP COLUMN IF EXISTS entitlement_source,
    DROP COLUMN IF EXISTS trial_ends_at,
    DROP COLUMN IF EXISTS trial_started_at;
