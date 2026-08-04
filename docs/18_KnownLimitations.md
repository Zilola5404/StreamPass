# StreamPass — Known Limitations

> Дата: 2026-08-03 | Только подтверждённые ограничения из кода

---

## Product Limitations

| Limitation | Detail |
|------------|--------|
| Single platform | Android only; no iOS, Windows, macOS |
| One-button connect | Android VPN: connect works; data-plane uses `VpnService.protect()` (v0.1.1+10) |
| No on-device routing | Decision Engine + rule hot-reload done; DIRECT/RELAY underlay sockets use `VpnService.protect()` (+10) |
| Exclusions local-only | Synced via GET/PUT `/api/v1/exclusions` (BL-014); local SharedPreferences cache remains |
| Boot auto-connect | «Автоподключение» при boot не поднимает VPN до открытия приложения (см. `BootReceiver.kt`) |
| No auto-update | Client APK: soft/hard prompt via GET /config (`latest_client_version` + `client_download_url`); rules/config already poll. Certs still manual. |
| Subscription cancel | Immediate revocation, not "stop auto-renewal" |

---

## Technical Limitations

| Limitation | Detail |
|------------|--------|
| Go monolith | No horizontal scaling without load balancer + multiple instances |
| Custom JWT/Redis | Not battle-tested at scale |
| Redis no persistence | Sessions lost on Redis restart |
| Health check | TCP probe only, not full Hysteria2 handshake |
| No go.sum | Reproducible builds not guaranteed |
| Admin via API key | No RBAC, no audit log |
| Telemetry no FK | Orphan events if user deleted |
| PostgreSQL sslmode=disable | Internal Docker network only |

---

## Infrastructure Limitations

| Limitation | Detail |
|------------|--------|
| Single VPS | No HA, no failover |
| nip.io domain | Not suitable for production branding |
| No CI/CD | Manual build and deploy |
| No monitoring | No Prometheus/Grafana |
| No automated backup | ~~Postgres volume only~~ → daily cron (BL-033) |
| Caddy single instance | No CDN, no WAF |

---

## Security Limitations

| Limitation | Detail |
|------------|--------|
| Debug signing (Android) | Release uses `key.properties` + JKS when present; otherwise debug fallback (BL-013) |
| connection_config plaintext | Relay secrets in DB without encryption |
| No refresh token rotation | Client stores refresh, no auto-refresh flow |
| Webhook no signature verify | Relies on re-fetch from provider |

---

## MVP Scope Exclusions (by design, ТЗ §21)

Kubernetes, ML, AI routing, MASQUE, Multipath QUIC, custom transport, ASN/GeoIP routing, browser extension, Linux desktop, OpenWRT, corporate version, referral system, multi-hop.

---

## Workarounds

| Limitation | Workaround |
|------------|------------|
| No Admin Panel | Use curl with X-Admin-Key |
| No CI/CD | Manual `go test ./...` before deploy |
| VPN stub | External Hiddify client with manual config (temporary) |
