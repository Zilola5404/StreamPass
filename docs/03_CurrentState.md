# StreamPass — Текущее состояние

> Дата: 2026-08-03 | Источник: анализ кодовой базы

---

## Что работает

### Backend (Go 1.22)

| Модуль | Описание | Файлы |
|--------|----------|-------|
| **Auth** | Register, Login, Logout. Argon2id, JWT access+refresh, Redis sessions | `backend/internal/application/auth/` |
| **Rule Service** | Версионированные правила. GET public, POST admin | `backend/internal/application/rule/` |
| **Config Service** | Динамическая конфигурация клиента | `backend/internal/application/configsvc/` |
| **Relay Manager** | Реестр relay, health-check, выбор лучшего | `backend/internal/application/relay/` |
| **Telemetry** | Приём метрик без PII | `backend/internal/application/telemetry/` |
| **Billing** | ЮKassa клиент, webhook, подписка | `backend/internal/application/billing/` |
| **Admin API** | ListUsers, CRUD relay, publish rules/config | `backend/internal/application/admin/` |
| **Health Monitor** | TCP-probe relay, POST health | `backend/cmd/healthmonitor/` |
| **Infrastructure** | Postgres migrations, Redis, rate limit, logging | `backend/internal/infrastructure/` |

### Docker Compose

Сервисы: `postgres`, `redis`, `backend`, `healthmonitor`, `caddy` — конфигурация в `docker-compose.yml`.

### Android Client (Flutter)

| Экран | Файл | Статус |
|-------|------|--------|
| Onboarding (login/register) | `client/lib/screens/onboarding_screen.dart` | ✅ |
| Home (connect orb) | `client/lib/screens/home_screen.dart` | ✅ UI |
| Servers | `client/lib/screens/servers_screen.dart` | ✅ |
| Subscription | `client/lib/screens/subscription_screen.dart` | ✅ |
| Settings | `client/lib/screens/settings_screen.dart` | ✅ |
| Exclusions | `client/lib/screens/exclusions_screen.dart` | ⚠️ Локально |
| Statistics | `client/lib/screens/statistics_screen.dart` | ✅ UI |
| Diagnostics | `client/lib/screens/diagnostics_screen.dart` | ✅ |

Сервисы: `auth_service.dart`, `streampass_api.dart`, `vpn_channel.dart`, `settings_service.dart`.

### Тесты (проходят)

- `shared/config` — YAML loader
- `backend/.../security` — Argon2 + JWT
- `backend/.../redisclient` — RESP parser
- `backend/.../middleware` — rate limiter
- `backend/.../rule`, `configsvc` — services
- `client/test/` — auth, widget, api_url

---

## Что реализовано частично

| Компонент | Что есть | Чего нет |
|-----------|----------|----------|
| **VPN Tunnel** | Android VPNService, TUN, TunnelBridge, Hysteria2, AAR, integration tests | Physical device TUN E2E (manual), Decision Engine |
| **Billing** | HTTP-клиент ЮKassa, webhook handler | Live-тест с реальным аккаунтом ЮKassa |
| **Exclusions** | UI экран | Синхронизация с backend, применение в Decision Engine |
| **Subscription cancel** | Немедленное прекращение доступа | Автопродление / отмена будущих списаний |
| **Scripts** | Build, RunTests, SmokeTest, Backup, GenerateDocs | Интеграция в CI/CD — TODO |
| **Документация** | `docs/`, `ai/`, `reports/`, `prompts/` | Заполнена, требует синхронизации при изменениях кода |

---

## Что не реализовано

| Компонент | По ТЗ | Статус |
|-----------|-------|--------|
| Decision Engine (клиент) | §5 | ❌ |
| Rule Engine (клиент) | §6 | ❌ |
| DNS Cache / DoH / DoQ | §7 | ❌ |
| Hysteria2 client (go_core) | §8 | ✅ Transport verified (BL-003 integration tests) |
| Fallback Strategy (UDP/TCP ports) | §10 | ❌ |
| Admin Panel UI | §11 | ❌ |
| Auto Update (клиент, правила, certs) | §16 | ❌ |
| Prometheus / Grafana | §18 | ❌ |
| CI/CD (GitHub Actions) | §19 | ❌ |
| Windows / macOS / iOS клиенты | §3 | ❌ |
| Client Core 90% shared code | §4 | ❌ |
| Decision Engine + Rule Engine | §5–6 | ❌ |

---

## Текущие проблемы

1. **BL-005/BL-006:** Decision Engine + Rule Engine на клиенте
2. **Physical Android VPN test** — manual when device + live relay available
3. **Production relay** — `212.43.157.167` (Hiddify Hysteria2 :32528); backend на `212.43.156.33`. См. `docs/RelayServers.md`

---

## Следующие шаги (приоритет)

1. **BL-005:** Decision Engine на клиенте
2. **Live-тест ЮKassa** — sandbox
3. **CI/CD** — GitHub Actions + VerifyBL003 in pipeline
4. **Physical Android E2E** — когда доступен device и online relay
