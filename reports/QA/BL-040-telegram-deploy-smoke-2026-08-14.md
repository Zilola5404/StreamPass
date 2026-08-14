# BL-040 Telegram deploy smoke — 2026-08-14

> VPS: `https://212-43-156-33.nip.io` · commit `d3e9a6d`  
> Bot: `@streampass_client_bot` (token set in server `.env` only)

## Deploy

| Step | Result |
|------|--------|
| `git pull` on VPS | ✅ `9560c60..d3e9a6d` |
| `docker compose up -d --build backend caddy` | ✅ |
| Migration 0011 (telegram payment cols) | ✅ auto on backend start |
| `TELEGRAM_BOT_TOKEN` + webhook secret in `.env` | ✅ (not in git) |
| APK OTA `downloads/StreamPass.apk` | ✅ +60 (14.0 MB) |

## Smoke

| Check | Result |
|-------|--------|
| `GET /health` | ✅ `{"status":"ok"}` |
| `GET /pay/` | ✅ 200 |
| Telegram `getWebhookInfo` | ✅ `…/api/v1/payments/telegram/webhook` |
| `POST /login` (test5) | ✅ access token |
| `GET /plans` | ✅ month/quarter/year in **XTR** (499/1299/3999) |
| `POST /payments` plan=month | ✅ `confirmation_url` → `https://t.me/…` invoice link |
| Backend log | ✅ `telegram invoice created` |
| `GET /payments/usdt/address` | ⏸ 503 — `USDT_TRC20_ADDRESS` not set |
| Webhook without secret | ✅ 403 FORBIDDEN (expected) |

## Manual (PO)

1. In Telegram: open `@streampass_client_bot` → `/buy` → pay test Stars invoice from app.
2. Optional: set `USDT_TRC20_ADDRESS` in VPS `.env` and restart backend.

## SLA (optional, not in this deploy)

- TUN connect ~6 s vs ≤5 s — unchanged; track as post-MVP polish.
