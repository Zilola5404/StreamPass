# StreamPass — Known Limitations

> Дата: 2026-08-05 | Только подтверждённые ограничения

---

## Product Limitations

| Limitation | Detail |
|------------|--------|
| Single platform | Android only; iOS / Windows / macOS — Open (BL-023…025) |
| Region coverage | Софт multi-region готов; в проде пока только NL relays |
| Boot auto-connect | «Автоподключение» при boot не поднимает VPN до открытия приложения |
| Subscription cancel | Immediate revocation; auto-renewal Blocked (нужна ЮKassa) |
| Cert rotation | Client soft/hard APK update есть; ротация TLS certs relay — manual |

---

## Technical Limitations

| Limitation | Detail |
|------------|--------|
| Go monolith | No HA without multiple instances + LB |
| Redis no persistence | Sessions lost on Redis restart |
| Health check | TCP probe only (не полный Hysteria2 handshake) |
| Admin via API key | `/admin/` + `X-Admin-Key`; нет RBAC / audit log |
| Telemetry no FK | Orphan events if user deleted |
| PostgreSQL sslmode=disable | Internal Docker network only |
| TCP underlay | UDP fallback портов есть; TCP underlay — later |

---

## Infrastructure Limitations

| Limitation | Detail |
|------------|--------|
| Single VPS | No HA / multi-AZ failover |
| nip.io domain | OK для MVP; не брендовый prod domain |
| Monitoring bind | Grafana/Prometheus на `127.0.0.1` (не публично) |
| Backup locality | Daily local gzip cron; off-site copy ещё нет |
| Caddy single instance | No CDN / WAF |

---

## Security Limitations

| Limitation | Detail |
|------------|--------|
| connection_config plaintext | Relay secrets in DB without field-level encryption |
| Webhook signature | ЮKassa path relies on provider re-fetch (live не проверен) |
| Admin key static | Rotation manual |

---

## MVP Scope Exclusions (ТЗ §21)

Kubernetes, ML, AI routing, MASQUE, Multipath QUIC, custom transport, ASN/GeoIP, browser extension, Linux desktop, OpenWRT, corporate, referral, multi-hop.

---

## Workarounds

| Limitation | Workaround |
|------------|------------|
| No DE/PL/FI VPS yet | Register new relays in Admin → Relays with region `de`/`pl`/`fi` |
| Off-site backup | Copy `/var/backups/streampass` to second host / S3 manually |
| Live ЮKassa | Sandbox keys + BL-004 when requested |
