# Telegram Stars + USDT payments (BL-040 / TZ 02.3)

## Env

```env
TELEGRAM_BOT_TOKEN=
TELEGRAM_WEBHOOK_SECRET=
STREAMPASS_PUBLIC_URL=https://212-43-156-33.nip.io
USDT_TRC20_ADDRESS=
```

When `TELEGRAM_BOT_TOKEN` is set, `GET /plans` returns Stars tariffs and `POST /payments` returns a Telegram `invoice_link`.

## Webhook

```bash
curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook" \
  -d "url=${STREAMPASS_PUBLIC_URL}/api/v1/payments/telegram/webhook" \
  -d "secret_token=$TELEGRAM_WEBHOOK_SECRET"
```

Endpoint: `POST /api/v1/payments/telegram/webhook`  
Header (optional): `X-Telegram-Bot-Api-Secret-Token: $TELEGRAM_WEBHOOK_SECRET`

## Bot UX

- `/buy` — lists Stars tariffs and points to the app / `/pay/`
- Stars payment → `pre_checkout_query` + `successful_payment` → subscription extended

## Public page

`https://<host>/pay/` — Stars helper + USDT address / confirm form (`web/pay.html`).

## App

Subscription screen opens `confirmation_url` / `invoice_link` via `url_launcher`.

## Notes

- YooKassa live keys are **not** required for MVP; Telegram path replaces them.
- USDT confirm is MVP-loose (TronGrid SUCCESS + unique `tx_hash`); tighten decoding later.
- Bot token is provided separately by PO; until set, backend keeps RUB/YooKassa plan path.
