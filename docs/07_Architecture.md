# StreamPass — Архитектура

> Дата: 2026-08-03 | Версия: 1.0

---

## 1. Общая схема системы

```mermaid
graph TB
    subgraph Client["Mobile Client (Flutter + Android VPN)"]
        UI[Flutter UI Screens]
        AuthSvc[Auth Service]
        APISvc[StreamPass API Client]
        VPNCh[VPN Channel]
        GoCore[Go Core - STUB]
        VPNService[Android VPNService]
        UI --> AuthSvc
        UI --> APISvc
        UI --> VPNCh
        VPNCh --> VPNService
        VPNService --> GoCore
    end

    subgraph Infrastructure["Docker Compose (VPS)"]
        Caddy[Caddy :443 TLS]
        Backend[Go Backend :8080]
        HM[Health Monitor]
        PG[(PostgreSQL 16)]
        Redis[(Redis 7)]
        Caddy --> Backend
        Backend --> PG
        Backend --> Redis
        HM --> Backend
    end

    subgraph Relay["Relay Servers"]
        R1[Hysteria2 Relay DE]
        R2[Hysteria2 Relay NL]
    end

    APISvc -->|HTTPS /api/v1| Caddy
    GoCore -.->|UDP Hysteria2| R1
    GoCore -.->|UDP Hysteria2| R2
    HM -->|TCP probe| R1
    HM -->|TCP probe| R2
```

> Пунктир: Go Core tunnel — stub, не реализован end-to-end.

---

## 2. Backend Architecture (Clean Architecture)

```mermaid
graph TB
    subgraph Delivery["Infrastructure / HTTP"]
        Router[router.go]
        Handlers[Handlers]
        MW[Middleware: Auth, RateLimit, Logging]
        Router --> Handlers
        MW --> Router
    end

    subgraph Application["Application Layer"]
        AuthApp[auth.Service]
        RuleApp[rule.Service]
        ConfigApp[configsvc.Service]
        RelayApp[relay.Service]
        TelemetryApp[telemetry.Service]
        BillingApp[billing.Service]
        AdminApp[admin.Service]
    end

    subgraph Domain["Domain Layer"]
        UserDom[user]
        RuleDom[rule]
        RelayDom[relay]
        ConfigDom[appconfig]
        SubDom[subscription]
        TeleDom[telemetry]
    end

    subgraph Infra["Infrastructure"]
        PGRepo[Postgres Repositories]
        RedisStore[Redis SessionStore]
        Security[Argon2 + JWT]
        YooKassa[YooKassa Client]
    end

    Handlers --> AuthApp
    Handlers --> RuleApp
    Handlers --> ConfigApp
    Handlers --> RelayApp
    Handlers --> TelemetryApp
    Handlers --> BillingApp
    Handlers --> AdminApp

    AuthApp --> UserDom
    RuleApp --> RuleDom
    ConfigApp --> ConfigDom
    RelayApp --> RelayDom
    BillingApp --> SubDom
    TelemetryApp --> TeleDom

    AuthApp --> PGRepo
    AuthApp --> RedisStore
    AuthApp --> Security
    BillingApp --> PGRepo
    BillingApp --> YooKassa
```

**Composition root:** `backend/cmd/server/main.go`

---

## 3. Frontend / Mobile Architecture

```mermaid
graph TB
    subgraph Flutter["Flutter App"]
        Main[main.dart]
        Screens[Screens]
        Services[Services]
        Main --> Screens
        Screens --> Services
    end

    subgraph Services
        AuthS[auth_service.dart]
        API[streampass_api.dart]
        VPN[vpn_channel.dart]
        Settings[settings_service.dart]
    end

    subgraph AndroidNative["Android Native"]
        MainAct[MainActivity.kt]
        VPNService[StreamPassVpnService.kt]
        Tunnel[TunnelBridge.kt]
        Boot[BootReceiver.kt]
        MainAct --> VPNService
        VPNService --> Tunnel
    end

    VPN --> MainAct
    AuthS --> SharedPrefs[(SharedPreferences)]
    API -->|HTTP| Backend
    Tunnel -.-> GoCoreAAR[streampasscore.aar STUB]
```

**Экраны:** onboarding, home, servers, subscription, settings, exclusions, statistics, diagnostics.

---

## 4. Database

```mermaid
erDiagram
    users ||--o{ payments : has
    users {
        text id PK
        text email UK
        text password_hash
        timestamptz subscription_active_until
    }
    payments {
        text id PK
        text user_id FK
        text provider_id UK
        bigint amount_rub
        int period_days
        text status
    }
    rule_sets {
        int version PK
        jsonb rules
    }
    app_configs {
        int version PK
        text min_supported_client_version
        boolean telemetry_enabled
    }
    relay_servers {
        text id PK
        text region
        text host
        int port
        boolean healthy
        text connection_config
    }
    telemetry_events {
        bigserial id PK
        text user_id
        int rtt_millis
        text relay_id
    }
```

Подробнее: `docs/09_Database.md`

---

## 5. Infrastructure

| Компонент | Технология | Конфиг |
|-----------|------------|--------|
| Reverse Proxy | Caddy 2 | `Caddyfile` |
| Backend | Go 1.22, distroless | `backend/Dockerfile` |
| Health Monitor | Go worker | `backend/cmd/healthmonitor/Dockerfile` |
| Database | PostgreSQL 16-alpine | `docker-compose.yml` |
| Cache | Redis 7-alpine | `docker-compose.yml` |
| TLS | Caddy auto-HTTPS | `212-43-156-33.nip.io` |

---

## 6. Взаимодействие сервисов

### Auth Flow
```
Client → POST /api/v1/register → Backend → Argon2id hash → Postgres
Client → POST /api/v1/login → Backend → JWT pair → Redis session
Client → GET /api/v1/servers → Bearer JWT → Backend validates → Postgres
```

### Relay Health Flow
```
Healthmonitor → GET /api/v1/servers/all (X-Admin-Key) → Backend
Healthmonitor → TCP probe each relay host:port
Healthmonitor → POST /api/v1/servers/health → Backend → Postgres update
Client → GET /api/v1/servers (Bearer) → filtered healthy relays
```

### Billing Flow
```
Client → POST /api/v1/payments (Bearer) → Backend → YooKassa CreatePayment
Client → opens confirmation_url in browser
YooKassa → POST /api/v1/payments/webhook → Backend → verify → extend subscription
Client → GET /api/v1/subscription → active_until
```

---

## 7. Planned Architecture (NOT IMPLEMENTED)

По ТЗ §4, но **не реализовано**:

- Client Core (Go) с Decision Engine, Rule Engine, DNS Cache
- Platform Adapters: WFP (Windows), Network Extension (macOS/iOS)
- 90% shared code across platforms
- Prometheus + Grafana monitoring

---

## 8. Ключевые файлы

| Файл | Назначение |
|------|------------|
| `backend/cmd/server/main.go` | Composition root |
| `backend/internal/infrastructure/http/router/router.go` | All routes |
| `backend/internal/infrastructure/postgres/migrations/` | DB schema |
| `client/lib/main.dart` | Flutter entry + API URL |
| `client/android/.../StreamPassVpnService.kt` | VPN TUN |
| `client/go_core/mobile/tunnel.go` | Tunnel stub |
| `docker-compose.yml` | Full stack |
| `shared/config/` | YAML config loader |
