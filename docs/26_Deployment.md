# StreamPass — Deployment

> Дата: 2026-08-05

---

## Production Deploy (Docker Compose)

### Prerequisites
- Linux VPS with Docker + Docker Compose
- Ports 80, 443 open
- Domain or nip.io address

### Steps

```bash
# 1. Clone
git clone <repo-url> streampass && cd streampass

# 2. Configure secrets
cp .env.example .env
# Edit .env: DB_PASSWORD, JWT_SECRET, ADMIN_API_KEY (required)
# Optional: YOOKASSA_*, REDIS_PASSWORD, GRAFANA_*

# 3. Deploy
docker compose up -d --build

# 4. Verify
curl http://localhost:8080/health
# Expected: {"status":"ok"}

# 5. Verify via Caddy (if domain configured)
curl https://212-43-156-33.nip.io/health
```

### Services Started

| Service | Internal Port | External Port |
|---------|---------------|---------------|
| postgres | 5432 | — (internal) |
| redis | 6379 | — (internal) |
| backend | 8080 | — (via Caddy) |
| healthmonitor | — | — |
| caddy | 80, 443 | 80, 443 |
| admin UI | — | via Caddy `/admin/` |
| prometheus | 9090 | local-only (127.0.0.1) |
| grafana | 3000 | local-only (127.0.0.1) |
| node-exporter | 9100 | internal |

**Prod:** `https://212-43-156-33.nip.io` · Admin: `https://212-43-156-33.nip.io/admin/`

---

## Local Development Deploy

```bash
# Backend only (requires local PG + Redis)
cd backend
cp config.example.yaml config.yaml
export DB_HOST=localhost DB_PASSWORD=... JWT_SECRET=... ADMIN_API_KEY=...
go run ./cmd/server

# Flutter client
cd client
flutter run --dart-define=STREAMPASS_API_URL=http://10.0.2.2:8080/api/v1
# 10.0.2.2 = Android emulator host
```

---

## Android Client Deploy

```bash
cd client
# BL-013 Done: place key.properties + JKS (not committed)
flutter build apk --release
# Output: build/app/outputs/flutter-apk/app-release.apk
# Current ship: StreamPass-v0.1.1+17-signed-arm64.apk
```

---

## Relay Server Deploy (External)

Hysteria2 relay deployed separately on Ubuntu VPS.  
Instructions in `docs/02_TZ.md` (relay setup section).

Not managed by StreamPass docker-compose.  
Prod: NL nodes; multi-region software ready (DE/PL/FI catalog).

---

## Update / Rollback

### Update
```bash
git pull
docker compose up -d --build
# Migrations auto-apply on backend startup
```

### Rollback
```bash
git checkout <previous-tag>
docker compose up -d --build
# Manual migration rollback if needed: run .down.sql
```

### Stop
```bash
docker compose down          # keep data
docker compose down -v       # delete postgres_data volume
```

---

## Caddy Configuration

**File:** `Caddyfile` — reverse_proxy backend; serves `admin/` at `/admin/`; blocks public `/metrics`.

For production domain, update Caddyfile and restart caddy service.

---

## Health Monitor

Runs automatically in docker-compose.  
Probes all relays every 60s via TCP connect.  
Reports to `POST /api/v1/servers/health`.

No separate deployment needed.

---

## Monitoring & Backups

| Component | Detail |
|-----------|--------|
| Prometheus | `infrastructure/prometheus/`; scrape backend metrics |
| Grafana | `infrastructure/grafana/`; `GRAFANA_ADMIN_*` in `.env` |
| Backups | Daily cron (BL-033) → `/var/backups/streampass`; off-site optional |

See `infrastructure/README.md`.

---

## CI/CD

**Status:** Done (BL-010) — `.github/workflows/ci.yml`

Workflow:
1. Push / PR → GitHub Actions
2. `go test ./...` + `flutter test`
3. Build checks as configured in workflow

Deploy to VPS remains manual (`docker compose up -d --build`).

---

## Monitoring Post-Deploy

| Check | Command |
|-------|---------|
| API health | `curl /health` |
| Admin UI | open `/admin/` |
| Auth | `curl -X POST /api/v1/login` |
| Smoke | `.\scripts\SmokeTest.ps1` |
| Docker logs | `docker compose logs -f backend` |
| HM logs | `docker compose logs -f healthmonitor` |
| Grafana | `http://127.0.0.1:3000` (local) |
| DB | `docker compose exec postgres psql -U streampass -c '\dt'` |

---

## Not Implemented

- Kubernetes deployment
- Blue-green deployment
- Automated rollback
- Secret manager integration
- Multi-region *hardware* (доп. VPS DE/PL/FI) — софт готов (каталог + picker); физические ноды добавляются через Admin
- Off-site backup copy (local daily Done)
