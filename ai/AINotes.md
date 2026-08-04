# AI Notes

> 2026-08-05

- VPN tunnel **не stub** — Hysteria2 + TUN + protect + Decision/DNS.
- `go.sum` в корне репо; CI в `.github/workflows/ci.yml`.
- Admin UI: `/admin/` + `ADMIN_API_KEY` из `.env` (не email/пароль).
- Monitoring: Prometheus/Grafana только на loopback.
- Backups: `scripts/backup-postgres.sh` + cron.
- Regions: API/каталог готовы; прод = NL.
- ЮKassa и Win/iOS/macOS — out of scope без запроса.
- APK: `StreamPass-v0.1.1+17-signed-arm64.apk`.
