# StreamPass — Changelog

> Формат: [Keep a Changelog](https://keepachangelog.com/)  
> Дата начала документа: 2026-08-03

---

## [Unreleased]

### Added
- Decision Engine + Rule Engine hot-reload on client (BL-005/BL-006, v0.1.1+6)
- `POST /api/v1/refresh` — access token renewal
- Connect diagnostics: `ConnectionLog`, `ConnectLogger`, Diagnostics screen export
- GitHub Actions CI: `go test`, `flutter test`

### Changed
- VPN connect flow: `PrepareRelay` before TUN (ADR-011)
- Relay ops docs: secrets replaced with placeholders in `docs/RelayServers.md`

### Fixed
- Startup crash: `StreamPassVpnService.onDestroy` no longer calls `stopSelf()`
- `pingMs` preserved on prepare-first connect path
- Concurrent `POST /refresh` deduplicated in `AuthService`

### Known Issues
- DIRECT routing on Android may need `VpnService.protect()`
- ЮKassa — not live-tested
- `connection_config` plaintext in PostgreSQL (BL/security backlog)

---

## [0.1.1] — 2026-08-04

### Added
- AI-friendly documentation structure (`docs/`, `ai/`, `reports/`, `prompts/`)
- Health Monitor worker (`backend/cmd/healthmonitor/`)
- Migration 0002: `connection_config` column on `relay_servers`
- Flutter Android client with screens: onboarding, home, servers, subscription, settings, exclusions, statistics, diagnostics
- Android VPNService scaffold with TunnelBridge

### Changed
- Relay handler: GET /servers now requires Bearer JWT (returns connection_config)
- README: note Health Monitor section is outdated (worker exists)

### Fixed
- Real relay data passed to VPN service (commit `50e9e15`)
- Build artifacts removed from VCS (commit `df531de`)

### Known Issues
- VPN tunnel (go_core) — stub, not functional
- ЮKassa — not live-tested
- No CI/CD

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
