# StreamPass — Changelog

> Формат: [Keep a Changelog](https://keepachangelog.com/)  
> Дата начала документа: 2026-08-03

---

## [Unreleased]

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
