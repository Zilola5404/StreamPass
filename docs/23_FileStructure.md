# StreamPass — File Structure

> Дата: 2026-08-05 | Actual repository layout

---

```
StreamPass/
├── README.md                          # Backend quick start
├── go.mod                             # Root Go module (streampass)
├── go.sum                             # Dependency checksums (BL-027)
├── docker-compose.yml                 # Full stack: PG, Redis, backend, HM, Caddy, Prom, Grafana
├── Caddyfile                          # Reverse proxy + /admin/
├── .env.example                       # Secrets template
├── .github/
│   └── workflows/ci.yml               # CI: go test + flutter test (BL-010)
│
├── shared/                            # Shared Go packages
│   ├── config/                        # YAML config loader + tests
│   ├── errors/                        # AppError codes
│   ├── logger/                        # Structured JSON logger
│   └── idgen/                         # ID generation
│
├── backend/
│   ├── cmd/
│   │   ├── server/main.go             # ★ Composition root, HTTP API
│   │   └── healthmonitor/             # Relay health worker
│   │       ├── main.go
│   │       └── Dockerfile
│   ├── internal/
│   │   ├── domain/                    # Domain models
│   │   ├── application/               # Use cases / services
│   │   ├── integrationtest/           # Postgres integration tests (BL-011)
│   │   └── infrastructure/
│   │       ├── postgres/              # Repositories + migrations
│   │       ├── redisclient/
│   │       ├── security/
│   │       ├── payment/yookassa/
│   │       ├── metrics/               # Prometheus metrics
│   │       └── http/
│   ├── config.example.yaml
│   └── Dockerfile
│
├── admin/                             # ★ Operator Admin UI (BL-020)
│   ├── index.html
│   ├── app.js
│   ├── styles.css
│   └── README.md
│
├── infrastructure/                    # ★ Monitoring configs (BL-021)
│   ├── prometheus/prometheus.yml
│   ├── grafana/provisioning/
│   ├── hysteria-test/
│   └── README.md
│
├── client/                            # Flutter mobile app
│   ├── lib/
│   │   ├── main.dart
│   │   ├── screens/
│   │   └── services/
│   ├── android/
│   │   └── app/src/main/kotlin/com/streampass/app/
│   │       ├── MainActivity.kt
│   │       ├── StreamPassVpnService.kt
│   │       ├── TunnelBridge.kt
│   │       ├── BootReceiver.kt
│   │       └── NativeSettingsChannel.kt
│   ├── go_core/                       # Go tunnel core (Hysteria2) — NOT stub
│   │   ├── mobile/                    # gomobile entry (tunnel, log)
│   │   ├── internal/
│   │   │   ├── decision/              # Decision Engine
│   │   │   ├── dnscache/              # DNS Cache + DoH
│   │   │   ├── hyconfig/              # hysteria2:// parse + fallback
│   │   │   ├── protect/               # VpnService.protect
│   │   │   └── tunbridge/             # sing-tun ↔ hysteria
│   │   ├── go.mod, go.sum
│   │   └── README.md
│   ├── android/app/libs/
│   │   └── streampasscore.aar
│   ├── test/                          # Flutter unit + e2e mock
│   └── pubspec.yaml
│
├── vendor-src/                        # Vendored Go modules
│   ├── crypto/
│   ├── sys/
│   ├── pq/
│   └── mobile/                        # golang.org/x/mobile (present)
│
├── docs/
├── ai/
├── reports/
├── prompts/
├── templates/
└── scripts/                           # SmokeTest, Backup, LoadTest, Build, …
```

---

## Key Entry Points

| Component | Entry File |
|-----------|------------|
| Backend API | `backend/cmd/server/main.go` |
| Health Monitor | `backend/cmd/healthmonitor/main.go` |
| Flutter App | `client/lib/main.dart` |
| VPN Service | `client/android/.../StreamPassVpnService.kt` |
| Tunnel Core | `client/go_core/mobile/tunnel.go` |
| Admin UI | `admin/index.html` |
| Routes | `backend/internal/infrastructure/http/router/router.go` |
| Migrations | `backend/internal/infrastructure/postgres/migrations/` |
| CI | `.github/workflows/ci.yml` |

---

## Not in Repository

| Item | Status |
|------|--------|
| iOS/Windows/macOS targets | Not found (BL-023…025 Open) |
| `client/android_old/` | Gitignored backup |
| `key.properties` / `*.jks` | Local only (BL-013; not committed) |
| `client/go_core/streampasscore.aar` | Prefer `android/app/libs/` |

---

## Documentation Structure

```
docs/
├── 00_ProjectRules.md     ← Rules
├── 01_Project.md          ← Passport
├── 02_TZ.md               ← Product spec
├── 03_CurrentState.md     ← Current state
├── 04_Backlog.md          ← Tasks
├── 05_Bugs.md             ← Bugs
├── 06_TestPlan.md         ← Testing
├── 07_Architecture.md     ← Architecture
├── 08_API.md              ← API
├── 09_Database.md         ← Database
├── 10_Progress.md         ← History
├── 11_Decisions.md        ← ADR
├── 12_ReleasePlan.md      ← Releases
├── 13_Risks.md            ← Risks
├── 14_AIContext.md        ← AI context
├── 15_DefinitionOfDone.md
├── 16_AI_HANDOFF.md
├── 17-32                    ← Extended docs
└── 99_ProjectDashboard.md
```
