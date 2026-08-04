# StreamPass — План тестирования

> Дата: 2026-08-05 | Версия: 1.1

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

**Дополнительно (go_core):** decision, hyconfig, dnscache, protect, tunbridge, tunnel tests.

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
| Integration | Full auth + connect flow | ✅ mock backend (VPN device E2E — manual recheck +17) |
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
