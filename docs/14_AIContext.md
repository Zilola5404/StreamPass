# StreamPass — AI Context

> **Самый важный файл для AI-агентов.** Прочитать первым после `00_ProjectRules.md`.  
> Дата: 2026-08-06 | Статус продукта синхронизирован с `docs/04_Backlog.md`

---

## Проект

**StreamPass** — система интеллектуальной маршрутизации интернет-трафика (**ускоритель**, не full-tunnel VPN).  
MVP: Go backend + Flutter Android client + Hysteria2 (foreign RELAY) + DIRECT для RU.

**Repo:** `C:\01_Projects\StreamPass` | **Branch:** `main`  
**Prod API:** `https://212-43-156-33.nip.io` | **Admin:** `/admin/`  
**Клиент:** `v0.1.1+23` (`ru-split-dns-v1`)

---

## Цель

Пользователь нажимает «Подключить» → система выбирает relay, применяет правила DIRECT/RELAY:  
российские сайты/приложения — **напрямую** (вне TUN / app-bypass), зарубежные (YouTube и т.п.) — через relay.

**Критично:** Domain DIRECT ≠ снятие `TRANSPORT_VPN`. См. `docs/33_DirectVsVpnBypass.md`.

---

## Архитектура (кратко)

```
Flutter Android UI (v0.1.1+23)
  → auth / API / VPN channel / rule engine / region picker
  → Android VpnService
      • RU IPv4 excludeRoute / intl-only routes (split-tunnel)
      • addDisallowedApplication (Госуслуги/ФНС/банки)
      • DNS: .ru → Yandex, foreign → DoH
  → TunnelBridge → go_core (Hysteria2 + decision + DNS)
  → HTTPS → Caddy → Go Backend :8080
  → PostgreSQL 16 + Redis 7
  → Hysteria2 relays (NL + region listeners de/pl/fi)
  → Health Monitor + Prometheus/Grafana (localhost)
  → Admin static UI (/admin/) + daily Postgres backups
```

Clean Architecture: `domain` → `application` → `infrastructure` → `http`  
Composition root: `backend/cmd/server/main.go`

Подробнее: `docs/07_Architecture.md`, `docs/33_DirectVsVpnBypass.md`

---

## Стек

| Layer | Tech |
|-------|------|
| Backend | Go 1.22, stdlib HTTP, lib/pq, custom JWT/Redis/YAML |
| DB | PostgreSQL 16, Redis 7 |
| Mobile | Flutter 3.x, Dart 3.3+, Kotlin VPNService |
| Tunnel | gomobile AAR, Hysteria2, Decision Engine, DNS Cache/DoH |
| Infra | Docker Compose, Caddy 2, distroless, Prometheus, Grafana |
| Payments | ЮKassa HTTP client (**live-тест пропущен**, BL-004) |

---

## Главные файлы

| Файл | Зачем |
|------|-------|
| `backend/cmd/server/main.go` | DI, startup, migrations |
| `backend/internal/infrastructure/http/router/router.go` | All API routes |
| `backend/internal/infrastructure/postgres/migrations/` | DB schema |
| `admin/` | Operator UI (X-Admin-Key) |
| `docker-compose.yml` | Full stack (backend, caddy, HM, prometheus, grafana) |
| `client/lib/main.dart` | Flutter entry, API URL |
| `client/lib/services/streampass_api.dart` | REST client |
| `client/go_core/mobile/tunnel.go` | Hysteria2 tunnel |
| `scripts/backup-postgres.sh` | Daily backups (BL-033) |
| `scripts/loadtest/` | API load test (BL-032) |
| `docs/04_Backlog.md` | Source of truth for task status |
| `docs/33_DirectVsVpnBypass.md` | DIRECT ≠ TRANSPORT_VPN; app-bypass + split DNS |
| `docs/18_KnownLimitations.md` | Подтверждённые ограничения |
| `ai/CurrentTask.md` | Current agent focus |

---

## Правила изменения кода

1. Не хардкодить секреты — только env / config placeholders  
2. API только под `/api/v1/`  
3. Admin endpoints — `X-Admin-Key`, constant-time compare  
4. Клиентские секреты relay — только после JWT  
5. После фичи: тесты + обновить `docs/04_Backlog.md` и `ai/CurrentTask.md`  
6. Не начинать BL-023/024/025 (Win/iOS/macOS) и live ЮKassa без явного запроса  

---

## Текущий статус (2026-08-05)

**Done:** VPN tunnel, Decision/Rule Engine, DNS/DoH, CI, Admin UI, monitoring, auto-update, regions, backups, Flutter E2E mock, API loadtest.  
**Blocked/Skipped:** ЮKassa live (BL-004), auto-renewal (BL-030).  
**Open by design:** Windows / iOS / macOS clients.  
**Ops residual:** physical device re-check +17; off-site backup copy; real domain.

Полный backlog: `docs/04_Backlog.md`
