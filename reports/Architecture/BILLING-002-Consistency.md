# BILLING-002 — Billing Consistency Fix

**Date:** 2026-09-04  
**Type:** small technical consistency commit (no new monetization features)

## 1. Trial hours SSOT

| Layer | Value |
|-------|--------|
| Config | `billing.trial_hours: 72` |
| Code | `auth.DefaultTrialHours = 72` |
| Wire | `main.go` → `WithTrialHours(cfg.IntOr("billing.trial_hours", DefaultTrialHours))` |
| Persistence | `trial_started_at` / `trial_ends_at` / `subscription_active_until` (server clock) |

`WithTrialDays` remains a **test helper** only (`days×24`). Production must use hours.

## 2. Plan codes (final)

Canonical MVP catalog (GET `/plans`, card/SBP mode):

| Code | Devices | Users |
|------|---------|-------|
| `personal_basic` | 2 | 1 |
| `personal_pro` | 5 | 1 |
| `business` | 5 | 5 |

Legacy aliases still accepted on CreatePayment via `NormalizePlanCode`:

`basic` / `month` / `year` → `personal_basic`  
`pro` → `personal_pro`

Alias rows removed from the wired catalog (no longer listed in GET `/plans`).

Telegram Stars period SKUs (`month` / `quarter` / `year`) remain when bot token is enabled.

## 3. Subscription statuses (official)

| Status | Connect | Meaning |
|--------|---------|---------|
| `TRIAL` | allowed | Free trial window |
| `ACTIVE` | allowed | Paid / admin entitlement |
| `CANCELED` | allowed until `active_until` | Auto-renew canceled; access continues to period end |
| `EXPIRED` | blocked | Had entitlement, now ended |
| `INACTIVE` | blocked | Never entitled |

`CANCELED` is an official **subscription** state (not a payment status).  
Payment provider `CANCELED` must not activate subscription.

`error_code=TRIAL_EXPIRED` is set **only** when `entitlement_source=trial` (paid expiry no longer mislabels).

## 4. Extra consistency fixes

- Login device limit uses plan_code (Pro/Business → 5; trial/basic → 2)
- `auth.max_devices` default in example config → `2`
- `GET /subscription` returns `max_devices`
- Constants in `domain/subscription/plans.go`

## Deploy

No new migration required (comments-only on `0012`). Redeploy backend + clients that read `max_devices` / statuses.
