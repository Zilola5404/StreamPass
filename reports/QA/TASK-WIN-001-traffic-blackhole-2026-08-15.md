# Windows traffic probe — 2026-08-15

## Root cause (confirmed live)

1. Leftover Wintun adapter **StreamPass** kept `0.0.0.0/0 via 10.10.0.2 metric 0` after a previous session.
2. `PhysicalInterfaceIndex` dialed through that route and returned **UNDERLAY_IF=StreamPass** (index 52).
3. Hysteria handshake + DIRECT sockets bound to TUN → timeouts / blackhole (sites and apps fail).
4. Relay was dialed **before** underlay bind (wrong order vs Android PrepareRelay).

## Evidence

- Before fix: `RELAY_CONNECT failed: ... timeout` + `UNDERLAY_IF ... name=StreamPass` + Access denied (non-admin).
- After fix: `UNDERLAY_IF index=17 name=Беспроводная сеть` + `RELAY_CONNECTED via=udp/... pingMs=152`.
- Stale default route removed → ya.ru / google / ifconfig.me OK again without TUN.

## Fixes shipped

- Skip StreamPass/Wintun/TAP/Outline when choosing underlay NIC.
- `BindPhysicalUnderlay` **before** Hysteria connect.
- Best-effort `route delete 0.0.0.0 mask 0.0.0.0 10.10.0.2` on connect.
- Clearer Admin error in Flutter engine.

## Operator

Run `streampass.exe` **as Administrator** (Wintun CreateAdapter). If sites die after crash:
`route delete 0.0.0.0 mask 0.0.0.0 10.10.0.2`
