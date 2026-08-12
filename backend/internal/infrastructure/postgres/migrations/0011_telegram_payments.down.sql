DROP INDEX IF EXISTS idx_payments_tx_hash;
ALTER TABLE payments
    DROP COLUMN IF EXISTS paid_at,
    DROP COLUMN IF EXISTS tx_hash,
    DROP COLUMN IF EXISTS tariff,
    DROP COLUMN IF EXISTS telegram_user_id,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS provider;
