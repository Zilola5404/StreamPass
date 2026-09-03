-- BILLING-001: orders + cancel flag + current plan code

CREATE TABLE IF NOT EXISTS orders (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_code     TEXT NOT NULL,
    amount_rub    BIGINT NOT NULL,
    period_days   INT NOT NULL,
    currency      TEXT NOT NULL DEFAULT 'RUB',
    status        TEXT NOT NULL DEFAULT 'PENDING',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS order_id TEXT REFERENCES orders(id);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS subscription_canceled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS plan_code TEXT;

COMMENT ON COLUMN users.subscription_canceled_at IS 'User canceled auto-renew; access until subscription_active_until';
COMMENT ON COLUMN users.plan_code IS 'Last confirmed plan code (personal_basic|personal_pro|business)';
