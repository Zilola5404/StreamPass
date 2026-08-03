# StreamPass — План тестирования

> Дата: 2026-08-03 | Версия: 1.0

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

**Не покрыто (TODO):**
- `application/auth` — register, login, logout
- `application/billing` — CreatePayment, webhook
- `application/relay` — ListAvailable, health
- `application/telemetry` — Record
- Postgres repositories — нужны integration tests

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

**Команда:**
```bash
cd client && flutter test
cd client && flutter analyze
```

---

## 2. Integration Tests

| Область | Описание | Статус |
|---------|----------|--------|
| Auth flow | Register → Login → JWT → Protected endpoint | ❌ Не реализовано |
| Billing flow | CreatePayment → Webhook → Subscription active | ❌ Не реализовано |
| Relay CRUD | Admin register → Health check → Client list | ❌ Не реализовано |
| Postgres repos | CRUD через testcontainers | ❌ Не реализовано |

**Рекомендация:** testcontainers-go или dockertest для Postgres/Redis.

---

## 3. API Tests

| Endpoint | Method | Auth | Тест | Статус |
|----------|--------|------|------|--------|
| /health | GET | — | 200 OK | Manual |
| /api/v1/register | POST | — | 201 + tokens | Manual |
| /api/v1/login | POST | — | 200 + tokens | Manual |
| /api/v1/rules | GET | — | 200 JSON | Manual |
| /api/v1/config | GET | — | 200 JSON | Manual |
| /api/v1/servers | GET | Bearer | 401 без token, 200 с token | Manual |
| /api/v1/telemetry | POST | Bearer | 204 | Manual |
| /api/v1/payments | POST | Bearer | 201 confirmation_url | Manual |
| /api/v1/subscription | GET | Bearer | 200 status | Manual |
| Admin endpoints | * | X-Admin-Key | 401 без key | Manual |

**Smoke test script:** `scripts/SmokeTest.ps1` (TODO: реализовать)

---

## 4. Security Tests

| Проверка | Описание | Статус |
|----------|----------|--------|
| Password hashing | Argon2id, не plaintext | ✅ Code review |
| JWT validation | Invalid/expired token → 401 | Manual |
| Rate limiting | Brute-force register/login | ✅ Unit test |
| Admin key | Constant-time compare | ✅ Code review |
| SQL injection | Parameterized queries (lib/pq) | ✅ Code review |
| Secrets in code | No hardcoded secrets | ✅ Code review |
| Telemetry PII | No URLs, no browsing history | ✅ Schema design |
| TLS | Caddy HTTPS termination | ✅ Docker Compose |

**TODO:**
- OWASP ZAP scan API
- Penetration test перед production

---

## 5. Load Tests

| Сценарий | Target | Статус |
|----------|--------|--------|
| GET /api/v1/rules | 100 RPS, p99 < 200ms | ❌ Не проводился |
| POST /api/v1/login | 20 RPS (rate limited) | ❌ |
| GET /api/v1/servers | 50 RPS authenticated | ❌ |

**Инструмент:** k6 или vegeta (TODO: BL-032)

---

## 6. Mobile Tests

| Тест | Описание | Статус |
|------|----------|--------|
| Widget tests | Onboarding, home screens | ✅ Базовые |
| Integration | Full auth + connect flow | ❌ |
| VPN permission | Android VPN permission dialog | Manual |
| Boot receiver | Autostart on boot | Manual |
| Subscription gate | Block connect without subscription | Manual |

---

## 7. CI Pipeline (TODO)

```yaml
# Планируемый GitHub Actions workflow
- go build ./...
- go vet ./...
- go test ./...
- cd client && flutter analyze && flutter test
- docker compose build
```

---

## 8. Критерии прохождения

| Уровень | Критерий |
|---------|----------|
| Unit | `go test ./...` green, `flutter test` green |
| Integration | Auth + Billing + Relay flows green |
| API | Smoke test all endpoints |
| Security | No critical findings |
| Load | p99 < 500ms at 50 RPS |
| Mobile | Connect flow E2E on real device |
