# StreamPass — Security Checklist

> Дата: 2026-08-03

---

## Authentication & Authorization

| Check | Status | Detail |
|-------|--------|--------|
| Password hashing | ✅ | Argon2id (`security/argon2_hasher.go`) |
| JWT access tokens | ✅ | HS256, 15m TTL, custom minimal impl |
| JWT refresh tokens | ✅ | Stored in Redis, 720h TTL, revocable |
| Admin endpoints | ✅ | X-Admin-Key, constant-time compare |
| Rate limiting | ✅ | 20 req/min on register/login/webhook |
| Brute-force protection | ✅ | Strict rate limit on auth endpoints |
| Input validation | ✅ | Email/password validation in auth service |

---

## Data Protection

| Check | Status | Detail |
|-------|--------|--------|
| Secrets in env only | ✅ | No hardcoded secrets in code |
| TLS in transit | ✅ | Caddy HTTPS termination |
| DB SSL | ⚠️ | sslmode=disable (Docker internal network) |
| connection_config at rest | ❌ | Plaintext in PostgreSQL |
| Telemetry no PII | ✅ | No URLs, no browsing history by design |
| Password not logged | ✅ | Structured logging excludes secrets |

---

## API Security

| Check | Status | Detail |
|-------|--------|--------|
| API versioning | ✅ | /api/v1/ prefix |
| Unified error format | ✅ | No stack traces to client |
| Auth on sensitive endpoints | ✅ | GET /servers requires Bearer |
| Webhook verification | ⚠️ | Re-fetches from provider, no signature check |
| CORS | N/A | Mobile client, no browser CORS needed |
| Request size limits | ⚠️ | Default Go http.Server limits |

---

## Infrastructure Security

| Check | Status | Detail |
|-------|--------|--------|
| Redis password | ✅ | requirepass in docker-compose |
| Postgres not exposed | ✅ | Internal Docker network only |
| Firewall | ⚠️ | Only ports 80/443 external; relay VPS separate |
| SSH keys only | TODO | Server hardening not documented |
| Fail2Ban | TODO | Not configured |
| Auto security updates | TODO | Not configured |
| Root login disabled | TODO | Not verified |

---

## Client Security

| Check | Status | Detail |
|-------|--------|--------|
| Token storage | ⚠️ | SharedPreferences (not encrypted) |
| Certificate pinning | ❌ | Not implemented |
| Release signing | ❌ | Debug keys (BUG-005) |
| VPN permission | ✅ | Standard Android VPN permission flow |
| ProGuard/R8 | TODO | Not verified for release |

---

## Pre-Production Checklist

- [ ] Production Android keystore
- [ ] Real domain with valid TLS cert
- [ ] Rotate JWT_SECRET and ADMIN_API_KEY
- [ ] Enable PostgreSQL SSL for external connections
- [ ] Encrypt connection_config at rest
- [ ] Implement webhook signature verification
- [ ] Security audit / penetration test
- [ ] OWASP ZAP scan on API
- [ ] Server hardening (SSH, Fail2Ban, UFW, auto-updates)
- [ ] Secret manager for production secrets
- [ ] Telemetry data retention policy
- [ ] Privacy policy and ToS

---

## Incident Response

TODO: Требуется уточнение

1. Revoke compromised JWT_SECRET → force re-login all users
2. Rotate ADMIN_API_KEY → update healthmonitor + operators
3. Revoke refresh tokens → Redis FLUSHDB (nuclear option)

---

## Compliance Notes

- Telemetry designed to exclude PII (ТЗ §14)
- No browsing history collection
- Subscription data stored in PostgreSQL
- TODO: GDPR/personal data policy if EU users
