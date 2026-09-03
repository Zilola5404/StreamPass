# Architect Release Gate — staged implementation

> Дата: 2026-08-25  
> Ветка: `fix/client-network-diagnostics`  
> Источник: замечания архитектора + Issues #2 / #4 + trial/Platega модель

---

## Уже закрыто ранее (код)

| Blocker | Статус |
|---------|--------|
| P0-1 Network Core (Issue #2) | Код готов; нужен device E2E evidence |
| P0-2 Server Unavailable UX (Issue #4) | Код готов; нужен device E2E |

---

## Этот этап (реализовано в коде)

### P0-3 / P0-4 — DIRECT ≠ обычный VPN (Windows)

Windows Wintun раньше ставил `0.0.0.0/0` без RU exclude → весь IPv4 в TUN → риск «как обычный VPN».

**Фикс:** `Inet4RouteExcludeAddress` из embedded `ru_ipv4_cidrs.txt` при `networkMode=split`  
Лог: `[vpn] split-tunnel mode=exclude-ru ruExcludes=N`

Android уже имел `excludeRoute` — без изменений.

### P0-5…P0-7 — Trial + Plans + PaymentProvider (Platega)

| Этап архитектора | Что сделано |
|------------------|-------------|
| A Trial Engine | Migration `0012`, register → **72h** `TRIAL` (`billing.trial_hours`), `GET /subscription` → `TRIAL`/`ACTIVE`/`CANCELED`/`EXPIRED`/`INACTIVE` |
| B Plan Catalog | `personal_basic` 299 / `personal_pro` 499 / `business` 1490 (aliases normalized; see BILLING-002) |
| C Payment abstraction | `default_provider: yookassa \| platega`, `Name()` на провайдерах |
| D Platega | `internal/infrastructure/payment/platega`, webhook `POST /payments/platega/webhook` |
| E/F Access | Client: trial banner + paywall; connect по-прежнему только при `isActive` (TRIAL\|ACTIVE) |
| G Recurring | **Не в этом PR** (после live Platega) |

**Важно:** Platega live credentials и compliance-подтверждение категории StreamPass — вне кода.  
Подписка активируется только после webhook CONFIRMED + `FetchPaymentStatus`.

### P0-10 Failover

Клиентский health poll: **60s → 10s** (ближе к SLA ≤10s ТЗ).

---

## Следующие этапы (не в этом коммите)

1. Device E2E matrix DIRECT/RELAY (ya.ru DIRECT / YouTube RELAY) на Android + Windows  
2. Live Platega payment (после MerchantId + письменного OK)  
3. Relay failover физически (выключить Relay A)  
4. Legal / monitoring alerts / restore test / branded domain  
5. Auto-renewal (Platega SubscriptionId)

---

## Конфиг

```yaml
billing:
  default_provider: platega   # or yookassa
  platega_merchant_id: ...
  platega_secret: ...
  platega_return_url: https://...
```

Миграция: `0012_trial_entitlement.up.sql` — прогнать на prod перед деплоем backend.
