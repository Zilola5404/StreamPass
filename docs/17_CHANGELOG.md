# StreamPass — Changelog

> Формат: [Keep a Changelog](https://keepachangelog.com/)  
> Дата начала документа: 2026-08-03 | Обновлено: 2026-08-25

---

## [Unreleased]

### Fixed
- **Issue #4:** Server Unavailable UX — HTTP timeouts, Home CTA «Подключить», русские ошибки, autoConnect limit, Windows TUN teardown on relay fail.
- **Android VPN notification:** тап по foreground/failure-уведомлению открывает `MainActivity` (`setContentIntent` + `PendingIntent`).
- **Android Connect / gVisor:** `streampasscore.aar` пересобран с `-tags with_gvisor` (ошибка `gVisor is not included in this build`). Команды сборки AAR в README обновлены.

### Added
- **TASK-WIN-001 / BL-023:** Windows TUN — Wintun sidecar (`streampasscore.exe`), physical-NIC `IP_UNICAST_IF` protect, DNS `198.18.0.1`, `traffic_ready` gate. Stages 5–12: Decision/split, Variant B IPv6, MTU UI, adaptive FALLBACK path, `scripts/VerifyWindowsTUN.ps1`.

### Docs
- `reports/QA/android-notification-gvisor-2026-08-25.md`

### Known Issues
- См. `docs/18_KnownIssues.md`
- TUN connect ~6s vs SLA ≤5s (accepted for Android MVP sign-off)
- T3 airplane recover: needs clean manual retest

---

## [0.1.1+55] — 2026-08-12

### Added
- **BL-040:** Telegram Stars + USDT payments (TZ 02.3) — webhook, `/buy`, `/pay/`, migration 0011
- **BL-052:** theme / language / About re-ship without SpAppBar overlays

### APK
- `StreamPass-v0.1.1+55-signed-arm64.apk`

---

## [0.1.1+54] — 2026-08-11

### Added
- Password show/hide on E01 + E10 (`PasswordField`)
- MVP human sign-off pack (payments → Telegram path; TUN ~6s noted)

### APK
- Build `StreamPass-v0.1.1+54-signed-arm64.apk` when ready

---

## [0.1.1+53] — 2026-08-11

### Added
- **BL-049:** устройства / лимит — `user_devices`, `GET|DELETE /me/devices`, login `device_id`, E10 список + «Отключить», `auth.max_devices` (default 3)
- **BL-051:** E05 «Уведомления о сбоях» + native alert на `error` / VPN revoke (не на ручной Disconnect)
- **BL-050:** Admin Premium/бан/поиск + `admin_audit_log` (`POST|DELETE /users/{id}/subscription`, ban, `GET /admin/audit`)

### APK
- `StreamPass-v0.1.1+53-signed-arm64.apk` (after build)

---

## [0.1.1+52] — 2026-08-11

### Fixed
- **BUG-IG-FEED:** must-relay TCP no longer uses 3s `relay_blackhole` kill — Instagram/Meta parallel TLS was false-positive’d (`i.instagram.com` dial ok → fail)

### Added
- BL-054 Terms/Privacy on E01 (from +51 tree)

### APK
- `StreamPass-v0.1.1+52-signed-arm64.apk`

---

## [0.1.1+51] — 2026-08-11

### Added
- **BL-054:** E01 register — «Условия» / «Политика конфиденциальности» (in-app screens)

### Changed
- Docs: KnownIssues / FinalAcceptance synced to +50 device pass; Instagram feed noted as BUG-IG-FEED

### APK
- `StreamPass-v0.1.1+51-signed-arm64.apk` (build when ready)

---

## [0.1.1+50] — 2026-08-10

### Fixed
- DNS `ExtractAIPs` now reads **Additional** (CNAME→A); 2ip.ru pin was empty → TCP wrongly RELAY
- Keep AAAA prefetch + PinDirectIP from +49

### APK
- `StreamPass-v0.1.1+50-signed-arm64.apk`

---

## [0.1.1+49] — 2026-08-10

### Fixed
- 2ip.ru IP-only TCP: AAAA-suppress now **prefetch A + PinDirectIP** (Android often never asks us for TypeA → host= empty → wrongly RELAY)

### APK
- `StreamPass-v0.1.1+49-signed-arm64.apk`

---

## [0.1.1+48] — 2026-08-10

### Fixed
- **ADR-018:** `DefaultMode=RELAY` again for foreign (ifconfig/Gemini/LinkedIn), plus **PinDirectIP** so `*.ru` / 2ip stay DIRECT (ISP IP)

### APK
- `StreamPass-v0.1.1+48-signed-arm64.apk`

---

## [0.1.1+47] — 2026-08-10

### Fixed
- **ADR-017:** rollback `DefaultMode=RELAY` → **DIRECT** (ТЗ: RU/2ip local; foreign must-relay + dial-fail→RELAY)
- Gemini: `gstatic.com` added to must-relay (assets); NetworkMonitor `connectivitycheck.gstatic.com` stays DIRECT

### APK
- `StreamPass-v0.1.1+47-signed-arm64.apk`

---

## [0.1.1+46] — 2026-08-10

### Changed
- **DefaultMode = RELAY** in product split (ADR-016): unmatched foreign on TUN exits via NL; RU stays DIRECT via excludeRoute + `*.ru` / NetworkMonitor
- Split now matches Full Relay for foreign sites (ifconfig / LinkedIn / Upwork / Indeed)

### APK
- `StreamPass-v0.1.1+46-signed-arm64.apk`

---

## [0.1.1+45] — 2026-08-10

### Fixed
- App bypass: `<queries>` + launcher discovery so Госуслуги/ФНС/S7 get `addDisallowedApplication` (was `appBypass=1` self-only under package visibility)
- Indeed/Upwork/LinkedIn: DNS-time RELAY IP pin + anycast HostsForIP so TCP no longer falls to `default_direct` (RU geo-block)

### APK
- `StreamPass-v0.1.1+45-signed-arm64.apk`

---

## [0.1.1+44] — 2026-08-10

### Fixed
- TUN stack `system` → **gvisor** (One UI: DNS worked, TCP never reached handler → Google/Upwork/LinkedIn/Gemini hung; RU sites OK via excludeRoute)

### APK
- `StreamPass-v0.1.1+44-signed-arm64.apk`

---

## [0.1.1+43] — 2026-08-10

### Fixed
- DNS→VPN: `addDnsServer(198.18.0.1)` instead of TUN `10.10.0.1` (One UI hairpin broke Chrome DNS)
- TCP DNS (port 53) answered by Go dnscache; DoT/853 still dropped
- Connect honors Diagnostics networkMode (`full_relay` / `direct_test` / `tcp_only`)
- NetworkMonitor hostnames resolve via fast Yandex/UDP first (VPN VALIDATED)

### APK
- `StreamPass-v0.1.1+43-signed-arm64.apk`

---

## [0.1.1+42] — 2026-08-09

### Fixed
- RELAY dial prefers `hostname:port` via HostForIP (Hiddify-like); VPS resolves A with `mode:4`
- Log: `[tun] relay-tcp rewrite` / `relay-udp rewrite`

### APK
- `StreamPass-v0.1.1+42-signed-arm64.apk`

---

## [0.1.1+41] — 2026-08-09

### Fixed
- One UI / Chrome: removed `allowFamily(AF_INET6)` — AAAA no longer dials outside VPN while tunnel sits idle
- `BuildInfo` / connect.log build label synced to ship number

### APK
- `StreamPass-v0.1.1+41-signed-arm64.apk`

---

## [0.1.1+40] — 2026-08-09

### Fixed
- VPS Hysteria IPv4-only: apps dialing AAAA → `no IPv4 address available` on relay
  - DNS: empty NOERROR for AAAA (`aaaa-suppress`)
  - TUN: drop native IPv6 destinations (`[tun] drop ipv6`)

### Evidence
- `reports/Audit/vps-hysteria-ipv6-block.md` (`journalctl -u hysteria`)

### APK
- `StreamPass-v0.1.1+40-signed-arm64.apk`

---

## [0.1.1+39] — 2026-08-09

### Added
- `[diag] reason=stream_open_no_data` (4s, 0 bytes) — traffic-path P0 audit Этап 0
- `direct_failed_retry_relay`: DIRECT fail → one marked RELAY retry (Этап 4; not in direct_test)
- `dnscache.ClassifyDNSRoute` / `SetRouteHint` — DNS route log follows Decision/forceMode

### Changed
- Default `[dns-route]` for unknown foreign: **DIRECT** (was misleading RELAY)

### APK
- `StreamPass-v0.1.1+39-signed-arm64.apk`
- Response: `reports/CodeReview/traffic-path-p0-audit-response.md`

---

## [0.1.1+36] — 2026-08-09

### Fixed
- Android «VPN без интернета»: NetworkMonitor hosts (`connectivitycheck.gstatic.com`, `clients3.google.com`, …) → **DIRECT** (не Google CIDR RELAY)
- Foreign DNS: DoH timeout 1.5s → fallback Yandex (Cloudflare часто медленный/блокируется в РФ)
- `VpnService.setMetered(false)` (API 29+)

### APK
- `StreamPass-v0.1.1+36-signed-arm64.apk`

---

## [0.1.1+35] — 2026-08-08

### Added
- `[vpn] traffic_ready` after first user-plane TCP/UDP byte (handshake ≠ traffic; Issue #1 / BL-001)
- Live integration: `TestIntegrationHysteriaUDPEcho` (DNS over Hysteria UDP)
- Architect Issue #1 stage progress: `reports/Architecture/ISSUE-1-stage-progress.md`

### Fixed
- IPv6 blackhole on IPv4-only VPN: `allowFamily(AF_INET6)` bypass
- D1: connection_config URI prefix no longer logged on bad scheme

### APK
- `StreamPass-v0.1.1+35-signed-arm64.apk` (`routing-policy-v1`)

---

## [0.1.1+34] — 2026-08-07

### Added
- Routing Policy SoT: `docs/07.4_RoutingPolicy.md`, ADR-015, TASK-02 Architecture Decision
- DNS-in-TUN (`vpn dns=10.10.0.1`) + `HostForIP` reverse map + `[conn]` / `[dns-route]` diagnostics
- Network Mode / MTU / UDP443 block in **Diagnostics (E09)** only (not E05 Settings)
- Operator QA tooling: `DiagnoseTrafficBlock.ps1`, `VerifyAppSiteSwitch.ps1`
- QA reports: `reports/QA/PRODUCT-QA-2026-08-07.md` + developer response

### Changed
- `DefaultMode=DIRECT` (FS §6 / 07.4); product path without `quic_direct_bypass` / silent RELAY→DIRECT
- FALLBACK only for `mode=FALLBACK`; must-relay fails with diag (blackhole 3s)
- Builtin + published rules: domain-first; Google/Meta CDN CIDR as **IP-only safety net**; no Cloudflare `/12`
- Accelerator rules API **v8**; client config `latest_client_version=0.1.1+34`
- OTA APK: `https://212-43-156-33.nip.io/downloads/StreamPass.apk` (`routing-policy-v1`)

### Fixed (QA PRODUCT-QA-2026-08-07)
- **BUG-001:** IP-only Meta/Google → DIRECT geo-block — CIDR RELAY safety net restored
- **BUG-003:** OTA/docs drift +25 vs codebase — ship **+34** + config/docs sync

### APK
- `StreamPass-v0.1.1+34-signed-arm64.apk` (`connectFlow = routing-policy-v1`)

---

## [0.1.1+25] — 2026-08-06 (prior ship)

### Added
- BL-042/043: password forgot/reset, GET/PUT/DELETE `/me`, Profile + ForgotPassword screens
- BL-048: GET `/plans`, GET `/payments`, month/year tariffs UI, soft cancel (access until `active_until`)
- BL-053: `SlaTargets` + `scripts/MeasureDeviceSLA.ps1`
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

### Known Issues (at +25)
- ЮKassa — **BL-040 Blocked** (нет live keys); client resume-poll готов
- BL-030 auto-renewal Blocked on BL-040
- Client connect/recover SLA — manual device pass (script covers API/cold launch)
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
