# Current Task

> Updated: 2026-08-04

## Главная задача

Деплой на сервер: `admin/` + monitoring stack (Prometheus/Grafana/node-exporter) + Caddyfile.

## Сделано локально

| Этап | Статус |
|------|--------|
| VPN +15 | APK готов; device E2E manual |
| BL-020 Admin Panel | Done (`/admin/`) |
| BL-021 Prometheus + Grafana | Done (compose + `/metrics` + dashboard) |

## Next

1. **Deploy** `admin/`, `Caddyfile`, `docker-compose.yml`, `infrastructure/` на `212.43.156.33`
2. Проверить `https://…/admin/` и Grafana `127.0.0.1:3000` (SSH tunnel)
3. Device E2E APK +15
4. BL-022 Auto Update **или** BL-027 go.sum hygiene
