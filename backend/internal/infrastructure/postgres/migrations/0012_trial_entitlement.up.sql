-- +0012_trial_and_entitlement_source
-- 3-day free trial on registration; entitlement source for TRIAL vs paid.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS trial_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS entitlement_source TEXT;

COMMENT ON COLUMN users.trial_started_at IS 'When free trial began (server-side SSOT)';
COMMENT ON COLUMN users.trial_ends_at IS 'When free trial ends; access via subscription_active_until';
COMMENT ON COLUMN users.entitlement_source IS 'trial | paid | admin';
