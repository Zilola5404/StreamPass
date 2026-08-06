# StreamPass — История разработки

> Формат: Дата | Что сделано | Файлы | Результат

---

| Дата | Что сделано | Файлы | Результат |
|------|-------------|-------|-----------|
| 2026-08-06 | FS SaaS: спецификация + сценарии + backlog BL-040…054 | `docs/02.2_*`, `34_*`, `35_*`, `04_Backlog` | Источник требований для Dev/QA |
| 2026-08-06 | BL-017 TCP underlay + device E2E script | `go_core/hyconfig`, `scripts/VerifyDeviceE2E.ps1`, VPS bridge | UDP→TCP fallback; APK +25 |
| 2026-08-06 | BL-035 off-site encrypted backup | `scripts/backup-offsite.sh`, cron 03:15 UTC | encrypt → `212.43.157.167` |
| 2026-08-05 | Docs sync to product reality | `docs/*`, `prompts/00_SystemPrompt.md` | VPN/admin/CI/monitoring reflected; dates 2026-08-05 |
| 2026-08-04 | BL-031/032: Flutter E2E mock + API loadtest | `client/test/e2e_flow_test.dart`, `scripts/loadtest` | E2E mock green; p99 baseline on prod |
| 2026-08-04 | BL-033: Postgres daily backup | `scripts/Backup.ps1`, cron, restore docs | `/var/backups/streampass` 03:00 UTC |
| 2026-08-04 | BL-026/027: multi-region + go.sum; APK +17 | regions API, picker, `go.sum` | NL in prod; DE/PL/FI software-ready |
| 2026-08-04 | BL-020/021/022: Admin UI, Grafana/Prometheus, OTA | `admin/`, `infrastructure/`, config API | `/admin/`, local metrics, soft/hard update |
| 2026-08-04 | BL-010…017: CI, integration tests, signing, DNS, exclusions, refresh, fallback | `.github/`, `integrationtest/`, go_core | P1 MVP items Done |
| 2026-08-04 | BL-005/BL-006: Decision Engine + rule hot-reload | `go_core/internal/decision/`, `RuleEngineService`, AAR | Per-flow DIRECT/RELAY; polling + UpdateRules |
| 2026-08-04 | Code review fixes (2c56972) | `tunnel.go`, `auth_service.dart`, docs | pingMs fix, refresh dedup, secrets redacted |
| 2026-08-04 | VPN connect + startup hardening | `tunnel.go`, `StreamPassVpnService.kt`, `auth/refresh` | PrepareRelay-first; POST /refresh; diagnostics log |
| 2026-08-03 | BL-003: VPN transport verification | integration tests, VerifyBL003.ps1, TunnelBridge fix | Hysteria connect + foreign IP PASS; APK OK |
| 2026-08-03 | BL-001/BL-002: Hysteria2 tunnel + AAR | `client/go_core/`, `streampasscore.aar` | go_core transport, Gradle AAR dep |
| 2026-08-03 | Инициализация AI-friendly документации | `docs/*`, `ai/*`, `reports/*`, `prompts/*` | Полная база знаний для AI-агентов |
| — | Первый релиз StreamPass MVP (backend) | `backend/`, `shared/`, `docker-compose.yml` | Backend API functional |
| — | Клиентская часть (Flutter Android) | `client/lib/`, `client/android/` | UI screens + VPNService scaffold |
| — | Связка backend с Android client | `client/lib/services/`, handlers | API integration working |
| — | Fix: real relay data in VPN service | `StreamPassVpnService.kt`, relay handler | Relay config passed to VPN |
| — | Health Monitor worker | `backend/cmd/healthmonitor/` | TCP probe + health reporting |
| — | Migration 0002: connection_config | `0002_relay_connection_config.up.sql` | Relay connection secrets in DB |
| — | Remove build artifacts from VCS | `.gitignore` | Cleaner repo |
| — | Update .gitignore | `.gitignore` | android_old, build dirs ignored |

---

## Git Commits (хронология, recent)

| Commit | Message |
|--------|---------|
| `8b5da33` | Flutter E2E на mock API (BL-031) и load test API (BL-032) |
| `67c4d9f` | Автоматический backup Postgres (BL-033) |
| `7ed1360` | Сборка клиента v0.1.1+17 с region picker (BL-026) |
| `e58c657` | Мульти-регионы relay (BL-026) и закрытие BL-027 |
| `1160899` | Автообновление клиента по config API (BL-022) |
| `391a64b` | Исправления VPN (+15), Admin Panel (BL-020) и мониторинг (BL-021) |
| `e272271` | Интеграционные тесты backend с Postgres (BL-011) |
| `9a9bfde` | Подпись release APK (BL-013) и DNS Cache + DoH (BL-016) |
| `31826d7` | CI/CD, README, UDP fallback портов и v0.1.1+11 |
| `bcfb491` | BL-005/BL-006: Decision Engine and client rule hot-reload |
| `86c6874` | BL-003: верификация VPN transport |
| `1d6ac76` | Документация + Hysteria2 tunnel |
| `6a6f029` | Связка бэкенд с клиентом Android |
| `9570b95` | Первый релиз StreamPass MVP |

### Backlog Done (summary, 2026-08-05)

DONE: BL-001..003, 005, 006, 010-017, 020-022, 026, 027, 031, 032, 033, 035  
SKIPPED: BL-004 YooKassa live  
BLOCKED: BL-030 (depends BL-004)  
OPEN intentional: BL-023 Windows, BL-024 iOS, BL-025 macOS

---

## Шаблон для новых записей

```
| YYYY-MM-DD | [Описание задачи] | [Изменённые файлы] | [Результат] |
```

Обновлять после каждой завершённой задачи.
