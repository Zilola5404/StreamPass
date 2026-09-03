ALTER TABLE users
    DROP COLUMN IF EXISTS plan_code,
    DROP COLUMN IF EXISTS subscription_canceled_at;

ALTER TABLE payments
    DROP COLUMN IF EXISTS order_id;

DROP TABLE IF EXISTS orders;
