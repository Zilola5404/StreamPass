# Monitoring (BL-021)

## Components

| Service | Role | Access |
|---------|------|--------|
| backend `/metrics` | HTTP counters + uptime | Docker network only (Caddy returns 404 publicly) |
| prometheus | scrape backend + node-exporter | `127.0.0.1:9090` |
| node-exporter | CPU / RAM / disk | internal |
| grafana | dashboards | `127.0.0.1:3000` |

## Env

```env
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=<strong-password>
```

Default password `changeme` — change before production use.

## SSH tunnel example

```bash
ssh -L 3000:127.0.0.1:3000 -L 9090:127.0.0.1:9090 root@212.43.156.33
```

Then open http://127.0.0.1:3000 (dashboard **StreamPass Overview**).
