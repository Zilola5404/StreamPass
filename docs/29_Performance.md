# StreamPass — Performance

> Дата: 2026-08-03

---

## Targets (ТЗ §22)

| Metric | Target | Current Status |
|--------|--------|----------------|
| Client startup | ≤ 2 seconds | TODO: Not measured |
| Connection establishment | ≤ 5 seconds | ❌ Tunnel stub — N/A |
| Auto-recovery after disconnect | ≤ 10 seconds | ❌ Not implemented |
| Server availability | ≥ 99.9% | TODO: Not measured |
| Rule update without reinstall | Required | ✅ API-based |

---

## Backend Performance

### Configuration

| Setting | Value | Location |
|---------|-------|----------|
| DB max_open_conns | 20 | config.example.yaml |
| DB max_idle_conns | 5 | config.example.yaml |
| DB conn_max_lifetime | 30m | config.example.yaml |
| Rate limit | 20 req/min | config.example.yaml |
| JWT access TTL | 15m | config.example.yaml |

### Design Decisions

| Decision | Rationale |
|----------|-----------|
| Custom Redis client | Minimal RESP2, only needed commands |
| Custom JWT | HS256 only, no heavy library |
| JSONB rules | Atomic publish, no JOIN overhead |
| Health check = liveness only | DB outage shouldn't kill all instances |
| Telemetry no FK | High-volume inserts without lock contention |

### Bottleneck Analysis (Theoretical)

| Component | Risk | Mitigation |
|-----------|------|------------|
| PostgreSQL | Single instance | Connection pool, indexes |
| Redis | Single instance | Session-only, low volume |
| Go monolith | Single process | Vertical scaling, future: multiple instances + LB |
| Rate limiter | In-memory per instance | Acceptable for single instance |

---

## Load Testing (TODO: BL-032)

| Scenario | Target | Status |
|----------|--------|--------|
| GET /api/v1/rules | 100 RPS, p99 < 200ms | Not tested |
| POST /api/v1/login | 20 RPS (rate limited) | Not tested |
| GET /api/v1/servers | 50 RPS authenticated | Not tested |
| POST /api/v1/telemetry | 100 RPS | Not tested |

**Tool:** k6 or vegeta (planned)

---

## Client Performance

| Area | Status |
|------|--------|
| Flutter startup | TODO: Not profiled |
| API polling intervals | Configurable via GET /config |
| VPN TUN setup | Native Android, should be fast |
| Memory usage | TODO: Not profiled |

---

## Relay Performance

| Metric | Source |
|--------|--------|
| RTT | Health monitor TCP probe → stored in relay_servers |
| Load ratio | Reported by health monitor |
| Bandwidth | Hysteria2 config: 1 gbps up/down (relay VPS) |

Client-side RTT measurement via telemetry POST.

---

## Monitoring (Not Implemented)

Planned (ТЗ §18):
- Prometheus metrics export
- Grafana dashboards
- CPU, RAM, relay load, active users, RTT, packet loss, error rate

---

## Optimization Opportunities

| Priority | Optimization | Impact |
|----------|-------------|--------|
| P1 | Implement VPN tunnel | Unblocks all client perf metrics |
| P2 | Add response caching for GET /rules, /config | Reduce DB reads |
| P2 | Connection pooling tuning under load | Stability |
| P3 | CDN for static assets (future web admin) | Latency |
| P3 | Multiple backend instances + LB | Horizontal scaling |

---

## Performance Testing Commands

```bash
# Backend build time
time go build ./...

# Test execution time
time go test ./...

# Simple load test (manual)
# hey -n 1000 -c 10 http://localhost:8080/api/v1/rules
# (requires 'hey' tool — not installed by default)
```
