# StreamPass — Environment

> Дата: 2026-08-05

---

## Environment Variables

**Template:** `.env.example` (copy to `.env`, never commit `.env`)

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_PASSWORD` | PostgreSQL password | Random string |
| `JWT_SECRET` | JWT signing key (32+ bytes) | `openssl rand -hex 32` |
| `ADMIN_API_KEY` | X-Admin-Key for admin endpoints | Random string |

### Database

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `postgres` (docker) / `localhost` (local) | PostgreSQL host |
| `DB_NAME` | `streampass` | Database name |
| `DB_USER` | `streampass` | Database user |
| `DB_PASSWORD` | — | **Required** |

### Redis

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ADDR` | `redis:6379` (docker) | Redis address |
| `REDIS_PASSWORD` | — | Redis password (required in docker-compose) |

### Payments (optional for dev)

| Variable | Description |
|----------|-------------|
| `YOOKASSA_SHOP_ID` | ЮKassa shop ID |
| `YOOKASSA_SECRET_KEY` | ЮKassa secret key |
| `YOOKASSA_RETURN_URL` | Payment return URL |

### Health Monitor

| Variable | Default | Description |
|----------|---------|-------------|
| `BACKEND_URL` | `http://backend:8080` | Backend URL for HM |
| `ADMIN_API_KEY` | (from .env) | Admin key for HM |
| `CHECK_INTERVAL` | `60s` | Probe interval |
| `CHECK_TIMEOUT` | `5s` | TCP probe timeout |

### Grafana / Monitoring

| Variable | Default | Description |
|----------|---------|-------------|
| `GRAFANA_ADMIN_USER` | `admin` | Grafana login user |
| `GRAFANA_ADMIN_PASSWORD` | `changeme` | Grafana login password (change in prod) |
| `GRAFANA_ROOT_URL` | `http://127.0.0.1:3000` | Grafana root URL (local-only) |

---

## Backend YAML Config

**Template:** `backend/config.example.yaml` → copy to `backend/config.yaml`

| Section | Key | Default | Description |
|---------|-----|---------|-------------|
| server | http_port | 8080 | API port |
| server | log_level | info | debug/info/warn/error |
| database | max_open_conns | 20 | Connection pool |
| database | max_idle_conns | 5 | Idle connections |
| database | conn_max_lifetime | 30m | Connection lifetime |
| jwt | access_ttl | 15m | Access token TTL |
| jwt | refresh_ttl | 720h | Refresh token TTL |
| rate_limit | public_requests_per_window | 20 | Rate limit count |
| rate_limit | public_window | 1m | Rate limit window |
| billing | plan_amount_rub | 299 | Subscription price |
| billing | plan_period_days | 30 | Subscription period |

All `${VAR}` placeholders resolved from environment at load time.

---

## Client Environment

| Variable | Location | Default |
|----------|----------|---------|
| `STREAMPASS_API_URL` | Compile-time (`--dart-define`) | `https://212-43-156-33.nip.io/api/v1` |

**Build with custom URL:**
```bash
flutter run --dart-define=STREAMPASS_API_URL=https://your-api.example/api/v1
```

---

## Development Setup

### Prerequisites
- Go 1.22+
- PostgreSQL 16
- Redis 7
- Flutter SDK (>=3.3.0)
- Docker + Docker Compose (for full stack)
- Android SDK (for mobile)

### Local Backend
```bash
cd backend
cp config.example.yaml config.yaml
export DB_HOST=localhost DB_NAME=streampass DB_USER=streampass \
       DB_PASSWORD=secret REDIS_ADDR=localhost:6379 \
       JWT_SECRET=$(openssl rand -hex 32) ADMIN_API_KEY=$(openssl rand -hex 32)
go run ./cmd/server
```

### Docker Full Stack
```bash
cp .env.example .env  # fill secrets
docker compose up -d --build
curl http://localhost:8080/health
```

### Flutter Client
```bash
cd client
flutter pub get
flutter run
```

---

## Production Environment

| Component | Current | Target |
|-----------|---------|--------|
| Domain | `212-43-156-33.nip.io` | `api.streampass.com` (TODO) |
| TLS | Caddy auto-HTTPS | Same |
| VPS | Single server | TODO: HA |
| Secrets | `.env` file | TODO: secret manager |
