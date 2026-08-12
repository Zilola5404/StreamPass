-- Telegram / USDT payment metadata (BL-040 redirect → TZ 02.3).
ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT 'yookassa',
    ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'RUB',
    ADD COLUMN IF NOT EXISTS telegram_user_id BIGINT,
    ADD COLUMN IF NOT EXISTS tariff TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS tx_hash TEXT,
    ADD COLUMN IF NOT EXISTS paid_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_tx_hash
    ON payments (tx_hash)
    WHERE tx_hash IS NOT NULL AND tx_hash <> '';
