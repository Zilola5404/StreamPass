# StreamPass — Known Issues

> Дата: 2026-08-11 | Клиент: **v0.1.1+50**  
> Связано: `docs/07.4_RoutingPolicy.md`, ADR-016…018

## Открытые (product / ops)

| ID | Приоритет | Описание | Статус |
|----|-----------|----------|--------|
| BUG-IG-FEED | P1 | Instagram лента: must-relay `relay_blackhole` (3s) убивал TLS | **Fixed +52** — blackhole only on FALLBACK / DIRECT→RELAY retry |
| Private DNS | P1 UX | Private DNS/DoT обходит VPN DNS → `host=` empty | **Off** required; +38 drops :853; +50 PinDirect mitigates known RU IPs |
| IPv6 | Note | VPN/relay IPv4-only; AAAA suppress | Until IPv6 egress on VPS |
| BL-040 | Blocked | ЮKassa live keys | No live billing |
| BL-054 | P2 | Terms / Privacy на E01 | **Done** +51 (in-app; public URL later) |

## Закрыто на +45…+50 (device adb)

| ID | Было | Статус |
|----|------|--------|
| Issue #1 / Stage 0 matrix | No adb / stale | **Pass on +50** — ifconfig NL, 2ip DIRECT, Gemini/LinkedIn RELAY, `appBypass=8` |
| BUG-001 geo IP-only | Foreign TCP `host=` → default DIRECT | **Mitigated** — DefaultMode RELAY + PinRelay/PinDirect + ExtractAIPs Unpack |
| BUG-002 Госуслуги VPN | `appBypass=1` | **Fixed** — `<queries>` + bypass list |
| 2ip shows NL | Pin empty / ExtractAIPs bug | **Fixed** +50 |

## Ограничения политики (не баги)

- `DefaultMode=DIRECT` (RELEASE-NETWORK-001) + DNS `PinDirectIP` for RU — unknown→ISP; known accelerator→RELAY.
- Product `split` **не** делает UDP/443→DIRECT (`quic_direct_bypass` запрещён).
- Cloudflare `/12` в builtin/rules **не** используется; Google/Meta CIDR — только IP-only safety net.
- Network Mode (full_relay / direct_test / tcp_only) — только **Диагностика (E09)**.

## Device checklist (+50)

1. Install `StreamPass-v0.1.1+50-signed-arm64.apk`
2. Private DNS = Off; Network Mode = Split; Connect
3. Log: `appBypass≥8`, `pin-direct` for `2ip.ru`, foreign `default_relay_foreign` / rule RELAY
4. ifconfig.me / ipify → relay IP; 2ip.ru → ISP; gov apps без VPN detection
5. Optional A/B: Diagnostics → `tcp_only` → Instagram feed
