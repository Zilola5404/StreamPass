# StreamPass — File Structure

> Дата: 2026-08-03 | Actual repository layout

---

```
StreamPass/
├── README.md                          # Backend quick start
├── go.mod                             # Root Go module (streampass)
├── docker-compose.yml                 # Full stack: PG, Redis, backend, HM, Caddy
├── Caddyfile                          # Reverse proxy config
├── .env.example                       # Secrets template
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
│   │   │   ├── user/
│   │   │   ├── rule/
│   │   │   ├── relay/
│   │   │   ├── appconfig/
│   │   │   ├── subscription/
│   │   │   └── telemetry/
│   │   ├── application/               # Use cases / services
│   │   │   ├── auth/
│   │   │   ├── billing/
│   │   │   ├── configsvc/
│   │   │   ├── relay/
│   │   │   ├── rule/
│   │   │   ├── telemetry/
│   │   │   └── admin/
│   │   └── infrastructure/
│   │       ├── postgres/              # Repositories + migrations
│   │       │   └── migrations/
│   │       │       ├── 0001_init.up.sql
│   │       │       └── 0002_relay_connection_config.up.sql
│   │       ├── redisclient/           # Custom Redis + SessionStore
│   │       ├── security/              # Argon2, JWT
│   │       ├── payment/yookassa/      # ЮKassa HTTP client
│   │       └── http/
│   │           ├── router/router.go   # ★ All routes
│   │           ├── handler/           # HTTP handlers
│   │           ├── middleware/        # Auth, rate limit, logging
│   │           └── response.go        # JSON helpers
│   ├── config.example.yaml
│   └── Dockerfile
│
├── client/                            # Flutter mobile app
│   ├── lib/
│   │   ├── main.dart                  # ★ Entry point, API URL
│   │   ├── screens/                   # UI screens (8)
│   │   └── services/                  # auth, api, vpn, settings
│   ├── android/
│   │   └── app/src/main/kotlin/com/streampass/app/
│   │       ├── MainActivity.kt
│   │       ├── StreamPassVpnService.kt  # ★ VPN TUN
│   │       ├── TunnelBridge.kt
│   │       ├── BootReceiver.kt
│   │       └── NativeSettingsChannel.kt
│   ├── go_core/                       # Go tunnel core (Hysteria2)
│   │   ├── mobile/tunnel.go           # ★ gomobile entry
│   │   ├── internal/hyconfig/         # hysteria2:// URI parser
│   │   ├── internal/tunbridge/        # sing-tun ↔ hysteria bridge
│   │   ├── go.mod, go.sum
│   │   └── README.md                  # gomobile build guide
│   ├── android/app/libs/
│   │   └── streampasscore.aar         # gomobile bind output
│   ├── test/                          # Flutter tests
│   └── pubspec.yaml
│
├── vendor-src/                        # Vendored Go modules
│   ├── crypto/                        # golang.org/x/crypto (argon2)
│   ├── sys/                           # golang.org/x/sys (transitive)
│   ├── pq/                            # github.com/lib/pq
│   └── mobile/                        # golang.org/x/mobile (gomobile)
│
├── docs/                              # ★ Project documentation (34 files)
├── ai/                                # ★ AI session state (8 files)
├── reports/                           # ★ Analysis reports
├── prompts/                           # ★ AI role prompts
├── templates/                         # Doc templates
└── scripts/                           # PowerShell automation
    ├── Build.ps1
    ├── RunTests.ps1
    ├── Backup.ps1
    ├── GenerateDocs.ps1
    └── SmokeTest.ps1
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
| Routes | `backend/internal/infrastructure/http/router/router.go` |
| Migrations | `backend/internal/infrastructure/postgres/migrations/` |

---

## Not in Repository

| Item | Status |
|------|--------|
| `.github/workflows/` | Not found |
| iOS/Windows/macOS targets | Not found |
| `client/android_old/` | Gitignored backup |
| `client/go_core/streampasscore.aar` | Build artifact (gitignored; use `android/app/libs/`) |

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
