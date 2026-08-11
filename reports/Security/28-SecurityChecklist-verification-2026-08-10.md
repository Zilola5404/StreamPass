# Security Checklist — Formal Verification Pass

> **Date:** 2026-08-10  
> **Source:** `docs/28_SecurityChecklist.md`  
> **Method:** Static code review + automated tests (no pentest / OWASP ZAP)

---

## Summary

| Area | PASS | PARTIAL | TODO / FAIL |
|------|------|---------|-------------|
| Authentication & Authorization | 7 | 0 | 0 |
| Data Protection | 4 | 1 | 1 |
| API Security | 4 | 2 | 0 |
| Infrastructure Security | 2 | 1 | 4 |
| Client Security | 4 | 1 | 2 |

**Overall:** Acceptable for **MVP / staging**. Pre-production items remain documented in checklist section "Pre-Production".

---

## Verified PASS (code + tests)

| Check | Evidence |
|-------|----------|
| Argon2id | `backend/internal/infrastructure/security/argon2_hasher.go` |
| JWT access 15m / refresh 720h Redis | `jwt_minimal.go`, `token_issuer.go`, `session_store.go` |
| Admin X-Admin-Key constant-time | `middleware/admin.go` |
| Rate limit 20/min auth | `middleware/ratelimit.go`, `router.go` |
| Input validation | `application/auth/validation.go` |
| No hardcoded prod secrets | `config.example.yaml`, env-only wiring |
| TLS via Caddy | `Caddyfile`, `docker-compose.yml` |
| Postgres internal only | no host port on postgres service |
| Telemetry no PII/URLs | `telemetry/service.go`, `tunbridge/bridge.go` |
| Connect log no secrets | `StreamPassVpnService.kt` (configLen only), `connect_flow_log_test.dart` |
| Release signing BL-013 | `build.gradle.kts`, `key.properties.example` |
| VPN permission | `AndroidManifest.xml`, `MainActivity.kt` |
| Auth integration test | `TestAuth_registerLoginRefreshLogout` |

**Tests run:** `go test ./internal/application/auth/ ./internal/infrastructure/security/ ./internal/infrastructure/http/middleware/` — all PASS.

---

## PARTIAL / known gaps (documented, not blockers for formal pass)

| Check | Status | Notes |
|-------|--------|-------|
| Token storage | PARTIAL | Primary: `flutter_secure_storage`; fallback to SharedPreferences if secure storage fails |
| DB SSL | PARTIAL | `sslmode=disable` OK inside Docker network |
| Webhook verification | PARTIAL | Re-fetch from provider; no signature check yet |
| Request size limits | PARTIAL | Default Go http.Server limits |
| Firewall / VPS hardening | PARTIAL | Caddy 80/443 only; SSH/Fail2Ban/UFW TODO |
| Certificate pinning | FAIL (planned) | Not implemented — pre-production |
| connection_config at rest | FAIL (planned) | Plaintext in PostgreSQL — pre-production |
| ProGuard/R8 | TODO | Not verified in this pass |

---

## Pre-Production (unchanged — out of scope for this pass)

- Real domain + valid TLS cert
- Rotate JWT_SECRET / ADMIN_API_KEY
- Encrypt connection_config
- Webhook signature verification
- Pentest / OWASP ZAP
- Server hardening (SSH keys, Fail2Ban, UFW, auto-updates)
- Secret manager, privacy policy, ToS

---

## Verdict

**Formal pass COMPLETE** for MVP checklist items verifiable without pentest.  
Checklist doc updated: token storage row, verification date.  
See also: `reports/SecurityReview.md` (initial static review).
