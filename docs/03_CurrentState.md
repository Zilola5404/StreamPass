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
| **VPN Tunnel** | Android VPNService, TUN, TunnelBridge, Hysteria2 client в go_core, streampasscore.aar | E2E тест на устройстве, Decision Engine |
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
| Hysteria2 client (go_core) | §8 | ⚠️ Реализован, нужен E2E тест на устройстве |
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

1. **VPN E2E не проверен на устройстве** — BL-003: нужен тест Connect + foreign IP (go_core и AAR готовы).
2. **ЮKassa не протестирована** против live/sandbox API.
3. **Нет CI/CD** — сборка и тесты только локально.
4. **Decision Engine / Rule Engine** на клиенте не реализованы — весь трафик идёт через relay.
5. **Release signing Android** — debug keys (TODO в `build.gradle.kts`).
6. **README backend** — статус Health Monitor может быть устаревшим; фактически worker есть в docker-compose.

---

## Следующие шаги (приоритет)

1. **BL-003: E2E VPN на Android** — Connect, проверка foreign IP через relay
2. **Decision Engine + Rule Engine на клиенте** — загрузка правил, маршрутизация DIRECT/RELAY
3. **Live-тест ЮKassa** — sandbox ключи, проверить CreatePayment/webhook
4. **CI/CD** — GitHub Actions: build, test, docker, gomobile (optional)
5. **Integration tests** — auth, billing, relay с testcontainers
6. **Обновить README** — синхронизировать статусы компонентов
