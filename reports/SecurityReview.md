# StreamPass — Security Review

> Date: 2026-08-03 | Reviewer: AI (initial, static analysis)

---

## Overall Assessment

**Rating: Acceptable for MVP dev stage.** Core security patterns implemented (Argon2id, JWT, rate limiting). Not ready for production without fixes.

---

## Findings

### Critical

| # | Finding | Location | Recommendation |
|---|---------|----------|----------------|
| S-01 | Android release uses debug signing keys | `build.gradle.kts` | Production keystore before distribution |
| S-02 | connection_config (relay secrets) stored plaintext | `relay_servers` table | Encrypt at rest or use secret manager |

### High

| # | Finding | Location | Recommendation |
|---|---------|----------|----------------|
| S-03 | Admin API key — single static secret | env ADMIN_API_KEY | RBAC via Admin Panel (post-MVP) |
| S-04 | No webhook signature verification | `billing_handler.go` | Verify ЮKassa signature |
| S-05 | Tokens in SharedPreferences (unencrypted) | `auth_service.dart` | Use flutter_secure_storage |

### Medium

| # | Finding | Location | Recommendation |
|---|---------|----------|----------------|
| S-06 | Custom JWT implementation | `jwt_minimal.go` | Add fuzz tests, consider mature library |
| S-07 | PostgreSQL sslmode=disable | config.yaml | OK for Docker internal; enable for external |
| S-08 | No certificate pinning on client | — | Add for production |
| S-09 | No server hardening documented | VPS | SSH keys, Fail2Ban, UFW per ТЗ §17 |

### Low / Informational

| # | Finding | Notes |
|---|---------|-------|
| S-10 | Health endpoint doesn't check deps | By design (liveness vs readiness) |
| S-11 | Telemetry user_id linkable | No PII collected, but linkable |
| S-12 | Rate limit in-memory per instance | OK for single instance |

---

## Positive Security Controls

- ✅ Argon2id password hashing
- ✅ JWT with short access TTL (15m)
- ✅ Refresh token revocation via Redis
- ✅ Rate limiting on auth endpoints
- ✅ Constant-time admin key comparison
- ✅ No secrets in source code
- ✅ TLS via Caddy
- ✅ Redis password required
- ✅ Telemetry schema excludes PII by design
- ✅ Input validation on auth (email/password)

---

## ТЗ §17 Compliance

| Requirement | Status |
|-------------|--------|
| TLS 1.3 | ✅ Caddy |
| Argon2id | ✅ |
| JWT | ✅ |
| SSH Keys Only | TODO |
| Disable Root Login | TODO |
| Fail2Ban | TODO |
| Firewall | ⚠️ Partial (Caddy ports only) |
| Auto Security Updates | TODO |
| Minimal logging | ✅ |

---

## Recommendations (Priority Order)

1. Production Android keystore (S-01)
2. Encrypt connection_config (S-02)
3. flutter_secure_storage for tokens (S-05)
4. Webhook signature verification (S-04)
5. Server hardening checklist (S-09)
6. Security audit before beta launch

---

## Not Tested

- No penetration testing performed
- No OWASP ZAP scan
- No dependency vulnerability scan
- Runtime behavior not verified

This is a static code review only.
