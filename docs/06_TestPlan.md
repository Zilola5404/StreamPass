# StreamPass — План тестирования

> Дата: 2026-08-06 | Версия: 1.2

---

## 1. Unit Tests

### Backend (Go)

| Пакет | Файл | Покрытие | Статус |
|-------|------|----------|--------|
| shared/config | `shared/config/config_test.go` | YAML loader | ✅ |
| security | `backend/.../security/security_test.go` | Argon2, JWT | ✅ |
| redisclient | `backend/.../redisclient/client_test.go` | RESP parser | ✅ |
| middleware | `backend/.../middleware/ratelimit_test.go` | Rate limiter | ✅ |
| rule service | `backend/.../rule/service_test.go` | Rule CRUD | ✅ |
| configsvc | `backend/.../configsvc/service_test.go` | Config CRUD | ✅ |
| router | `backend/.../router/router_test.go` | v1 prefix helper | ✅ |
| metrics | `backend/.../metrics/metrics_test.go` | Prometheus helpers | ✅ |

**Дополнительно (go_core):** decision, hyconfig (incl. TCP underlay), dnscache, protect, tunbridge, mobile/tunnel tests.

**Последний прогон (2026-08-06):** `go test ./...` backend PASS; `go test ./...` in `client/go_core` PASS; `flutter test` **49/49** PASS.

**Команда:**
```bash
go test ./...
go vet ./...
```

### Client (Flutter/Dart)

| Файл | Scope | Статус |
|------|-------|--------|
| `client/test/auth_service_test.dart` | Offline login error | ✅ |
| `client/test/widget_test.dart` | Onboarding when logged out | ✅ |
| `client/test/api_url_test.dart` | /api/v1 URL prefix | ✅ |
| `client/test/connect_flow_log_test.dart` | VPN connect log / invalid relay | ✅ |
| `client/test/e2e_flow_test.dart` | Login → Home → Regions (mock) | ✅ BL-031 |

**Команда:**
```bash
cd client && flutter test
cd client && flutter analyze
```

---

## 2. Integration Tests

| Область | Описание | Статус |
|---------|----------|--------|
| Auth flow | Register → Login → JWT → Protected endpoint | ✅ BL-011 `backend/internal/integrationtest/` |
| Billing flow | CreatePayment → Webhook → Subscription active | ✅ BL-011 (mocked provider as designed) |
| Relay CRUD | Admin register → Health check → Client list | ✅ BL-011 |
| Postgres repos | CRUD via real Postgres in harness | ✅ BL-011 |

**Расположение:** `backend/internal/integrationtest/` (harness + api_test).

---

## 3. API Tests

| Endpoint | Method | Auth | Тест | Статус |
|----------|--------|------|------|--------|
| /health | GET | — | 200 OK | ✅ SmokeTest |
| /api/v1/register | POST | — | 201 + tokens | Manual / integration |
| /api/v1/login | POST | — | 200 + tokens | Manual / integration |
| /api/v1/rules | GET | — | 200 JSON | ✅ SmokeTest |
| /api/v1/config | GET | — | 200 JSON | ✅ SmokeTest |
| /api/v1/regions | GET | — | 200 JSON | ✅ SmokeTest |
| /api/v1/servers | GET | Bearer | 401 без token, 200 с token | Manual / integration |
| /api/v1/telemetry | POST | Bearer | 204 | Manual |
| /api/v1/payments | POST | Bearer | 201 confirmation_url | Manual (live Skipped) |
| /api/v1/subscription | GET | Bearer | 200 status | Manual |
| Admin endpoints | * | X-Admin-Key | 401 без key | SmokeTest optional `-AdminKey` |

**Smoke test script:** `scripts/SmokeTest.ps1` — **implemented** (prod default `https://212-43-156-33.nip.io`).

---

## 4. Security Tests

| Проверка | Описание | Статус |
|----------|----------|--------|
| Password hashing | Argon2id, не plaintext | ✅ Code review |
| JWT validation | Invalid/expired token → 401 | Manual / integration |
| Rate limiting | Brute-force register/login | ✅ Unit test |
| Admin key | Constant-time compare | ✅ Code review |
| SQL injection | Parameterized queries (lib/pq) | ✅ Code review |
| Secrets in code | No hardcoded secrets | ✅ Code review |
| Telemetry PII | No URLs, no browsing history | ✅ Schema design |
| TLS | Caddy HTTPS termination | ✅ Docker Compose |
| Release signing | key.properties path | ✅ BL-013 |

**TODO:**
- OWASP ZAP scan API
- Penetration test перед production

---

## 5. Load Tests

| Сценарий | Target | Статус |
|----------|--------|--------|
| GET /health, /rules, /config, /regions | p99 < 500ms @ ~25 RPS | ✅ BL-032 (`scripts/loadtest`) |
| POST /api/v1/login | 20 RPS (rate limited) | Manual / optional `-email` |
| GET /api/v1/servers | authenticated via LoadTest.ps1 | Optional |

**Инструмент:** `go run ./scripts/loadtest` / `.\scripts\LoadTest.ps1` / optional k6 (`scripts/loadtest/k6-public.js`)

---

## 6. Mobile Tests

| Тест | Описание | Статус |
|------|----------|--------|
| Widget tests | Onboarding, home screens | ✅ Базовые |
| E2E (mock API) | Login → Home → Regions picker | ✅ BL-031 `test/e2e_flow_test.dart` |
| Integration | Full auth + connect flow | ✅ mock backend; device E2E API — `VerifyDeviceE2E.ps1` (+25, 2026-08-06) |
| Off-site backup | Cron + `.enc` on secondary | `VerifyOffsiteBackup.ps1` (needs SSH) |
| VPN permission | Android VPN permission dialog | Manual |
| Boot receiver | Autostart on boot | Manual |
| Subscription gate | Block connect without subscription | Manual |

---

## 7. CI Pipeline (Done — BL-010)

`.github/workflows/ci.yml`:
- `go build` / `go vet` / `go test`
- `flutter analyze` / `flutter test`

```bash
# Local parity
go test ./...
cd client && flutter analyze && flutter test
```

---

## 8. Критерии прохождения

| Уровень | Критерий |
|---------|----------|
| Unit | `go test ./...` green, `flutter test` green |
| Integration | Auth + Billing + Relay flows green (BL-011) |
| API | SmokeTest endpoints green |
| Security | No critical findings |
| Load | p99 < 500ms baseline (BL-032) |
| Mobile | Connect flow E2E on real device (recheck pending) |
| **Traffic behavior** | `VerifyTrafficBehavior.ps1` + manual site/app matrix |

---

## 9. Traffic & lifecycle tests (функциональность продукта)

Проверяет **как работает трафик** (сайт/приложение → DIRECT/RELAY/bypass), **запуск**, **отключение без краша**.

### Автоматические (CI / локально)

| Область | Файл / команда | Что проверяет |
|---------|----------------|---------------|
| Decision matrix | `client/go_core/internal/decision/traffic_matrix_test.go` | yandex.ru DIRECT, youtube RELAY, … |
| DNS matrix | `client/go_core/internal/dnscache/traffic_dns_test.go` | `.ru` → Yandex, foreign → DoH |
| gomobile API | `client/go_core/mobile/traffic_matrix_test.go` | `DecideRoute()` контракт |
| Flutter contract | `client/test/traffic_behavior_test.dart` | rules JSON, bypass list, subscription gate |
| VPN lifecycle | `client/test/vpn_lifecycle_test.dart` | connect→disconnect state, logging |
| App startup | `client/test/app_startup_test.dart` | cold start без краша |

```bash
# Go traffic matrix
cd client/go_core && go test ./... -run TrafficMatrix -count=1

# Flutter
cd client && flutter test test/traffic_behavior_test.dart test/vpn_lifecycle_test.dart
```

### Device + logcat (adb)

| Скрипт | Описание |
|--------|----------|
| `scripts/VerifyTrafficBehavior.ps1` | Unit matrix + manual checklist + log patterns |
| `scripts/VerifyTrafficBehavior.ps1 -AfterConnect` | Logcat: app-bypass, split-tunnel, relay connected |
| `scripts/VerifyTrafficBehavior.ps1 -AfterTrafficRu` | VPN profile DNS 77.88.8.8 + split-tunnel (RU via OS) |
| `scripts/VerifyTrafficBehavior.ps1 -AfterTraffic` | Go `[dns] via=doh` after opening youtube.com |
| `scripts/VerifyTrafficBehavior.ps1 -AfterDisconnect` | Logcat: stop complete, no FATAL |
| `scripts/CollectConnectLogs.ps1` | Сбор connect.log / logcat |
| `scripts/traffic_expectations.json` | Матрица сайтов/приложений и ожидаемых log patterns |

### Manual matrix (на телефоне)

| Категория | Ожидание |
|-----------|----------|
| RU сайты (yandex.ru, 2ip.ru) | RU IP, DIRECT, DNS Yandex |
| Foreign (YouTube, Instagram) | Через relay (ускорение) |
| Госуслуги / банки (native) | Открываются без «отключите VPN» (app bypass) |
| Chrome | RU через split DNS+CIDR; foreign через TUN |
| Disconnect | UI «Готов к подключению», **приложение не закрывается** |
| APK update / revoke | `onDestroy` без crash (см. `StreamPassVpnService.kt`) |

См. также `docs/33_DirectVsVpnBypass.md`.
