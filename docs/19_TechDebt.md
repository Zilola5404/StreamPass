# StreamPass — Technical Debt

> Дата: 2026-08-03

---

## Critical (P0)

| ID | Debt | Location | Effort |
|----|------|----------|--------|
| TD-001 | VPN tunnel stub — no real Hysteria2 | `client/go_core/mobile/tunnel.go` | Large |
| TD-002 | No integration tests | backend application layer | Medium |
| TD-003 | ЮKassa never live-tested | `backend/.../payment/yookassa/` | Small |
| TD-004 | README outdated (Health Monitor) | `README.md` | Trivial |

---

## High (P1)

| ID | Debt | Location | Effort |
|----|------|----------|--------|
| TD-010 | No CI/CD pipeline | `.github/workflows/` missing | Medium |
| TD-011 | Missing go.sum | repo root | Small |
| TD-012 | vendor-src/mobile missing | `client/go_core/go.mod` | Medium |
| TD-013 | Auth service untested | `backend/.../auth/` | Medium |
| TD-014 | Billing service untested | `backend/.../billing/` | Medium |
| TD-015 | Postgres repos untested | `backend/.../postgres/` | Large |
| TD-016 | Android debug signing | `client/android/app/build.gradle.kts` | Small |
| TD-017 | Empty PowerShell scripts | `scripts/*.ps1` | Small |

---

## Medium (P2)

| ID | Debt | Location | Effort |
|----|------|----------|--------|
| TD-020 | No refresh token auto-rotation (client) | `client/lib/services/auth_service.dart` | Medium |
| TD-021 | Exclusions not synced | `client/lib/screens/exclusions_screen.dart` | Medium |
| TD-022 | Custom JWT instead of library | `backend/.../security/jwt_minimal.go` | Low (accept) |
| TD-023 | Custom Redis instead of go-redis | `backend/.../redisclient/` | Low (accept) |
| TD-024 | Custom YAML instead of gopkg.in/yaml | `shared/config/` | Low (accept) |
| TD-025 | android_old backup dir | `client/android_old/` | Trivial |
| TD-026 | Client README is Flutter template | `client/README.md` | Small |
| TD-027 | Subscription cancel semantics | `backend/.../billing/` | Small |

---

## Low (P3)

| ID | Debt | Location | Effort |
|----|------|----------|--------|
| TD-030 | No telemetry retention policy | DB + application | Small |
| TD-031 | No connection_config encryption | `relay_servers` table | Medium |
| TD-032 | Health handler doesn't check DB/Redis | `health_handler.go` | Small (by design) |
| TD-033 | Hardcoded default API URL in client | `client/lib/main.dart` | Small |

---

## Accepted Debt (documented, not fixing)

| Item | Rationale (ADR) |
|------|-----------------|
| Custom JWT/Redis/YAML | ADR-002: dependency-free infrastructure |
| stdlib router | ADR-003: Go 1.22+ ServeMux sufficient |
| JSONB rules | ADR-005: MVP scale |
| Admin API key | ADR-004: no Admin Panel in MVP |

---

## Debt Paydown Priority

1. TD-001 (VPN tunnel) — unblocks MVP
2. TD-010 + TD-011 + TD-012 — unblocks reliable development
3. TD-002 + TD-013 + TD-014 + TD-015 — quality gate
4. TD-003 — unblocks monetization
5. Everything else — post-MVP
