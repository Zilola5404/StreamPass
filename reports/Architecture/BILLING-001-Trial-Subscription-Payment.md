# BILLING-001 — Trial → Subscription → Payment Architecture

**Date:** 2026-09-03  
**Priority:** P0 monetization MVP  
**Status:** Implemented (code + unit tests). Release Gates 5–8 remain operational/legal.

## Goal

Single access model owned by StreamPass Backend:

| Status | Connect |
|--------|---------|
| `TRIAL` | allowed |
| `ACTIVE` | allowed |
| `CANCELED` (until `active_until`) | allowed — **official subscription state (BILLING-002)** |
| `EXPIRED` | blocked |
| `INACTIVE` | blocked |

App UI is never locked; only Connect is gated.

## What shipped

### Trial (72h SSOT)

- Registration sets `trial_started_at`, `trial_ends_at`, `subscription_active_until`, `entitlement_source=trial`.
- Duration: **`billing.trial_hours` (default 72)** — wall-clock hours on server, not device clock.
- Example: register 01 Sep 14:00 → trial ends 04 Sep 14:00.

### Access policy

- `GET /subscription` returns `access_allowed`, `hours_left`, `error_code` (`TRIAL_EXPIRED`), `plan_code`, statuses including `CANCELED`.
- Client uses `access_allowed` for Connect; expired trial opens Subscription / paywall copy.

### Plan Catalog (Backend SSOT)

Canonical codes (prices from config, not Flutter):

| Code | Default RUB | Devices | Users |
|------|-------------|---------|-------|
| `personal_basic` | 299 | 2 | 1 |
| `personal_pro` | 499 | 5 | 1 |
| `business` | 1490 | 5 | 5 |

`GET /plans` returns canonical plans only. Aliases (`basic`/`pro`/`month`/`year`) still accepted by `POST /payments`.

### Orders + payments

- Migration `0013_billing001_orders`: `orders` table, `payments.order_id`, `users.plan_code`, `users.subscription_canceled_at`.
- Flow: **Select plan → Create Order → Create Payment → Provider URL → Webhook → CONFIRMED → Subscription ACTIVE**.
- Endpoints: `POST /orders`, `POST /payments` (`plan_code` and/or `order_id`).
- **No activation** on success URL or client button.

### Idempotency

- `MarkSucceededIfPending` / `MarkSucceededByIDIfPending`: `UPDATE … WHERE status='PENDING'`.
- Replay webhook → no second `ActivatePaidPlan`.
- Non-`CONFIRMED` / canceled provider status → no activation.

### Payment providers

Abstraction unchanged: YooKassa / Platega / Telegram Stars / USDT. Existing providers kept.

### Client

- Removed hardcoded plan price fallbacks.
- Paywall: «Пробный период закончился» / «Выберите тариф, чтобы продолжить пользоваться StreamPass.» / CTA «Выбрать тариф».
- Connect blocked after trial → opens subscription screen; app remains usable.

## Definition of Done (engineering)

| Check | Evidence |
|-------|----------|
| New user gets Trial | register + `trial_hours=72` |
| Trial = 72h server clock | unit + register use case |
| Device clock cannot extend trial | server `trial_ends_at` only |
| Connect blocked after trial | `access_allowed=false`, `TRIAL_EXPIRED` |
| App not blocked | home still loads; paywall banner |
| Plan Catalog from Backend | `GET /plans` |
| Payment → webhook → ACTIVE | `HandleWebhook` + tests |
| Duplicate webhook safe | `TestWebhookIdempotency_NoDuplicateActivate` |
| Canceled payment no ACTIVE | `TestCanceledPaymentDoesNotActivate` |

## Release Gates (not closed by this PR)

5. **Real first payment** — needs production-like Platega/YooKassa payment + live webhook.  
6. **Production server** — VPS restart policy, monitoring, backups, Relay A/B.  
7. **Security** — secrets audit, JWT/webhook validation review.  
8. **Legal** — Privacy / ToS / Refund / support contacts.

## Deploy notes

1. Apply migration `0013_billing001_orders`.
2. Set `billing.trial_hours: 72` and plan RUB keys in config.
3. Redeploy backend before shipping client that expects `access_allowed` / `personal_*` codes.
