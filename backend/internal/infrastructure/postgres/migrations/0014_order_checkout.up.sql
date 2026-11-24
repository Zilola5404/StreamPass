ALTER TABLE orders
    ADD COLUMN checkout_started BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN checkout_url TEXT NOT NULL DEFAULT '';

-- Existing unpaid orders with provider payments must not initiate another charge.
UPDATE orders
SET checkout_started = TRUE
WHERE status = 'PENDING'
  AND EXISTS (SELECT 1 FROM payments WHERE payments.order_id = orders.id);
