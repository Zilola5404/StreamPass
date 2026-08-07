# StreamPass — Known Issues

> Дата: 2026-08-08 | Клиент: **v0.1.1+35** (`routing-policy-v1`)  
> Связано: Issue #1, `docs/07.4_RoutingPolicy.md`, BL-001 re-validation

## Открытые (product / ops)

| ID | Приоритет | Описание | Статус |
|----|-----------|----------|--------|
| Issue #1 / E2E | P0 | Physical Stage 0 matrix на устройстве | **Blocked** — нет adb; APK +35 локально |
| BUG-001 | P0→fix pending retest | Foreign geo-block IP-only | CIDR safety net +34; retest on +35 |
| BUG-002 | P1 | Госуслуги VPN visibility | Bypass in code; verify on device |
| IPv6 | Note | VPN IPv4-only; AF_INET6 bypass outside TUN (+35) | Documented product choice until IPv6 TUN |
| BL-040 | Blocked | ЮKassa live keys | No live billing |

## Ограничения политики (не баги)

- `DefaultMode=DIRECT` — неизвестный destination не уходит в RELAY (FS §6 / 07.4).
- Product `split` **не** делает UDP/443→DIRECT (`quic_direct_bypass` запрещён).
- Cloudflare `/12` в builtin/rules **не** используется; Google/Meta CIDR — только IP-only safety net.
- Network Mode (full_relay / direct_test / tcp_only) — только **Диагностика (E09)**.

## Device checklist после OTA +35

1. Install local `StreamPass-v0.1.1+35-signed-arm64.apk` (OTA APK upload may lag config)
2. Private DNS = Off; Network Mode = Split; reconnect
3. Connect log: `vpn dns=10.10.0.1`, `build=0.1.1+35`, `ipv6=bypass`, `[vpn] traffic_ready`
4. ya.ru / 2ip.ru DIRECT; YouTube RELAY; Госуслуги bypass
5. Config API already: `latest_client_version=0.1.1+35`
