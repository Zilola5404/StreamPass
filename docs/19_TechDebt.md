# StreamPass — Technical Debt

> Дата: 2026-08-05

---

## Critical (P0)

| ID | Debt | Location | Effort |
|----|------|----------|--------|
| TD-001 | VPN tunnel stub — no real Hysteria2 | `client/go_core/mobile/tunnel.go` | Done (BL-001…003) |
| TD-002 | No integration tests | backend application layer | Done (BL-011) |
| TD-003 | ЮKassa never live-tested | `backend/.../payment/yookassa/` | Small (BL-004 Skipped) |
| TD-004 | README outdated (Health Monitor) | `README.md` | Done (BL-012) |

---

## High (P1)

| ID | Debt | Location | Effort |
|----|------|----------|--------|
| TD-010 | No CI/CD pipeline | `.github/workflows/` | Done (BL-010) |
| TD-011 | Missing go.sum | repo root | Done (BL-027) |
| TD-012 | vendor-src/mobile missing | `client/go_core/go.mod` | Done |
| TD-013 | Auth service untested | `backend/.../auth/` | Done (BL-011 integration) |
| TD-014 | Billing service untested | `backend/.../billing/` | Done (BL-011 integration) |
| TD-015 | Postgres repos untested | `backend/.../postgres/` | Done (BL-011 integration) |
| TD-016 | Android debug signing | `client/android/app/build.gradle.kts` | Done (BL-013) |
| TD-017 | Empty / incomplete PowerShell scripts | `scripts/*.ps1` | Small (частично: SmokeTest, Backup, LoadTest Done) |

---

## Medium (P2)

| ID | Debt | Location | Effort |
|----|------|----------|--------|
| TD-020 | No refresh token auto-rotation (client) | `client/lib/services/auth_service.dart` | Done (BL-015) |
| TD-021 | Exclusions not synced | `client/lib/screens/exclusions_screen.dart` | Done (BL-014) |
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
| Admin API key | ADR-004: Admin Panel UI exists; API key still used for HM/ops |

---

## Debt Paydown Priority

1. TD-003 — ЮKassa live (only on explicit request; unblocks BL-030)
2. TD-027 — subscription cancel semantics
3. TD-026 / TD-017 — docs/scripts polish
4. TD-030 / TD-031 — retention + encrypt connection_config
5. Everything else — post-MVP / accepted
