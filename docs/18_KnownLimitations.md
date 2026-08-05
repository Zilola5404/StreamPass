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
| TCP underlay | Done: framed TCP→UDP bridge на VPS (`streampass-tcpunderlay`, TCP 8443/24443 → UDP 443). TCP/443 занят Caddy |
| Android VPN key icon | `VpnService` всегда показывает системный ключ; снять флаг у приложения можно только через `addDisallowedApplication` |
| Domain DIRECT ≠ OS bypass | Decision Engine DIRECT не снимает `TRANSPORT_VPN`; для Госуслуг/ФНС/банков нужен app-bypass |
| KindAPP rules | Backend Kind=APP пока нет (не в ТЗ §6); список пакетов в клиенте + эвристика по установленным приложениям |
| Browser under VPN | Chrome нельзя целиком disallow (иначе foreign sites не ускорятся); RU-сайты → split DNS + RU CIDR exclude |
| Split DNS | `.ru/.su/.рф` → Yandex `77.88.8.8`; foreign → Cloudflare DoH (`dnscache`) |

См. также: `docs/33_DirectVsVpnBypass.md`.

---

## Infrastructure Limitations

| Limitation | Detail |
|------------|--------|
| Single VPS | No HA / multi-AZ failover |
| nip.io domain | OK для MVP; не брендовый prod domain |
| Monitoring bind | Grafana/Prometheus на `127.0.0.1` (не публично) |
| Backup locality | Daily dump на primary; encrypted off-site на `212.43.157.167` (BL-035) |
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
| Off-site backup | Done: `scripts/backup-offsite.sh` → second host + `PullBackupsOffsite.ps1` |
| Live ЮKassa | Sandbox keys + BL-004 when requested |
