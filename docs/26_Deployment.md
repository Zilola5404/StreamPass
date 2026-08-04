# StreamPass — Deployment

> Дата: 2026-08-03

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
# Optional: YOOKASSA_*, REDIS_PASSWORD

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
flutter build apk --release
# Output: build/app/outputs/flutter-apk/app-release.apk
# NOTE: Currently uses debug signing keys (TODO: BL-013)
```

---

## Relay Server Deploy (External)

Hysteria2 relay deployed separately on Ubuntu VPS.  
Instructions in `docs/02_TZ.md` (relay setup section).

Not managed by StreamPass docker-compose.

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

**File:** `Caddyfile`
```
212-43-156-33.nip.io {
    reverse_proxy backend:8080
}
```

For production domain, update Caddyfile and restart caddy service.

---

## Health Monitor

Runs automatically in docker-compose.  
Probes all relays every 60s via TCP connect.  
Reports to `POST /api/v1/servers/health`.

No separate deployment needed.

---

## CI/CD

**Status:** Not implemented.

Planned workflow:
1. Push to main → GitHub Actions
2. `go test ./...` + `flutter test`
3. `docker compose build`
4. Push images to registry
5. Deploy to VPS (manual or automated)

See `docs/04_Backlog.md` BL-010.

---

## Monitoring Post-Deploy

| Check | Command |
|-------|---------|
| API health | `curl /health` |
| Auth | `curl -X POST /api/v1/login` |
| Docker logs | `docker compose logs -f backend` |
| HM logs | `docker compose logs -f healthmonitor` |
| DB | `docker compose exec postgres psql -U streampass -c '\dt'` |

---

## Not Implemented

- Kubernetes deployment
- Blue-green deployment
- Automated rollback
- Secret manager integration
- Multi-region *hardware* (доп. VPS DE/PL/FI) — софт готов (каталог + picker); физические ноды добавляются через Admin
