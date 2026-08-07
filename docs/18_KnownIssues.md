# StreamPass — Known Issues

> Дата: 2026-08-07 | Клиент: **v0.1.1+34** (`routing-policy-v1`)  
> Связано: `reports/QA/PRODUCT-QA-2026-08-07.md`, `docs/07.4_RoutingPolicy.md`

## Открытые (product / ops)

| ID | Приоритет | Описание | Статус |
|----|-----------|----------|--------|
| BUG-001 | P0→fix pending retest | Foreign sites geo-block when IP-only + DIRECT | CIDR safety net in **+34** / rules v8 — **device retest** |
| BUG-002 | P1 | Госуслуги «отключите VPN» | Bypass packages in code; verify log `VPN app-bypass:` after +34 reconnect |
| BUG-004 | P2 | UI «подписка неактивна» при network error | Mitigated in client (`_subscriptionCheckFailed`) — QA retest |
| BUG-005 | P2 | BL-035 off-site backup proof on secondary VPS | Scripts exist; cron/`.enc` not fully verified by QA |
| BL-040 | Blocked | ЮKassa live keys | No live billing |
| BL-030 | Blocked | Auto-renewal | Depends on BL-040 |
| — | Open | Windows / iOS / macOS clients | BL-023…025 |
| — | Backlog | `connection_config` plaintext in Postgres | Security |

## Ограничения политики (не баги)

- `DefaultMode=DIRECT` — неизвестный destination не уходит в RELAY (FS §6 / 07.4).
- Product `split` **не** делает UDP/443→DIRECT (`quic_direct_bypass` запрещён).
- Cloudflare `/12` в builtin/rules **не** используется; Google/Meta CIDR — только IP-only safety net.
- Network Mode (full_relay / direct_test / tcp_only) — только **Диагностика (E09)**.

## Device checklist после OTA +34

1. Install `https://212-43-156-33.nip.io/downloads/StreamPass.apk`
2. Private DNS = Off; Network Mode = Split; reconnect
3. Connect log: `vpn dns=10.10.0.1`, `build=0.1.1+34`
4. YouTube / Instagram / Gemini: `[decision] action=RELAY`
5. Госуслуги: без VPN block; log `VPN app-bypass:`

```powershell
.\scripts\DiagnoseTrafficBlock.ps1 -LiveProbe -ReportPath reports\QA\traffic-block-retest.md
.\scripts\VerifyAppSiteSwitch.ps1 -AutoLaunch -SkipManual -ReportPath reports\QA\traffic-switch-retest.md
```
