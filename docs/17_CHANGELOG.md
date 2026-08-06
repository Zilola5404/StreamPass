# StreamPass — Changelog

> Формат: [Keep a Changelog](https://keepachangelog.com/)  
> Дата начала документа: 2026-08-03 | Обновлено: 2026-08-06

---

## [Unreleased]

### Added
- TCP underlay fallback (BL-017): client framed UDP-over-TCP + VPS `streampass-tcpunderlay` (TCP 8443/24443 → Hysteria UDP 443)
- Device E2E script checks TCP underlay ports + config/download/rules (`scripts/VerifyDeviceE2E.ps1`)
- Off-site encrypted Postgres backups to second host (BL-035)
- Admin Panel UI at `/admin/` (BL-020)
- Prometheus + Grafana local monitoring (BL-021)
- Client auto-update via config API (BL-022)
- Multi-region catalog + region picker; prod NL nodes (BL-026)
- `go.sum` for reproducible builds (BL-027)
- Flutter E2E mock flow (BL-031); API loadtest scripts (BL-032)
- Daily Postgres backup cron (BL-033)
- APK **v0.1.1+25** signed arm64 (`connectFlow = tcp-underlay-v1`)
- Audit remediation: secure token storage, webhook secret, ADR-012/013/014

### Changed
- Docs aligned to product reality (2026-08-05): VPN not stub; CI/admin/monitoring present

### Fixed
- (see 0.1.1 notes and git log for VPN crash / protect / sideload fixes)

### Known Issues
- ЮKassa — not live-tested (BL-004 Skipped)
- BL-030 auto-renewal Blocked on BL-004
- Client T1–T4 perf not formally measured on device
- `connection_config` plaintext in PostgreSQL (security backlog)
- Windows / iOS / macOS clients Open (BL-023…025)

---

## [0.1.1] — 2026-08-04

### Added
- Hysteria2 VPN tunnel in go_core + streampasscore.aar (BL-001…003) — **not a stub**
- Decision Engine + Rule Engine hot-reload on client (BL-005/BL-006)
- DNS Cache + DoH (BL-016); UDP port fallback (BL-017)
- Exclusions sync (BL-014); refresh token rotation (BL-015)
- Release signing via `key.properties` (BL-013)
- GitHub Actions CI: `go test`, `flutter test` (BL-010)
- Backend integration tests with Postgres (BL-011)
- README refresh (BL-012)
- Admin UI, Grafana/Prometheus, OTA, regions, backups, loadtest, E2E (BL-020…022,026,027,031–033)
- `POST /api/v1/refresh`; connect diagnostics; regions API
- AI-friendly documentation structure (`docs/`, `ai/`, `reports/`, `prompts/`)
- Health Monitor worker (`backend/cmd/healthmonitor/`)
- Migration 0002+: `connection_config`, exclusions, auto-update, region normalize

### Changed
- VPN connect flow: `PrepareRelay` before TUN (ADR-013)
- Relay handler: GET /servers requires Bearer JWT (returns connection_config)
- Client version bumps through **v0.1.1+17**

### Fixed
- Real relay data passed to VPN service
- Startup / disconnect / protect / native-lib sideload crashes
- `pingMs` preserved on prepare-first connect path
- Concurrent `POST /refresh` deduplicated in `AuthService`

### Known Issues (at 0.1.1 cut)
- ЮKassa — not live-tested
- Measured client perf targets pending

---

## [0.1.0] — First MVP Release

### Added
- Go backend with Clean Architecture
- Auth: register, login, logout (Argon2id, JWT, Redis sessions)
- Rule Service: versioned rules (GET/POST)
- Config Service: versioned client config (GET/POST)
- Relay Manager: server registry, health ingestion
- Telemetry: metrics ingestion without PII
- Billing: ЮKassa HTTP client, webhook, subscription
- Admin API: X-Admin-Key protected endpoints
- PostgreSQL migrations (auto-apply)
- Docker Compose: postgres, redis, backend, caddy
- Rate limiting, structured logging, unified errors

---

## Template

```markdown
## [X.Y.Z] — YYYY-MM-DD

### Added
### Changed
### Fixed
### Removed
### Security
```
