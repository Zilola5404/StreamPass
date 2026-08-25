# Issue #2 — P0 Network Core Reliability: implementation report

> Date: 2026-08-25  
> Source: https://github.com/Zilola5404/StreamPass/issues/2  
> Branch: `fix/client-network-diagnostics`

## Summary

Implemented code-level remediations for idle disconnect, stale routes, recovery, and CONNECTED ≠ TRAFFIC_READY. Physical-device E2E evidence (idle 5/15/30 min, Doze, sleep/wake matrices) remains **required for DoD** and is **not** claimed complete.

---

## Checklist (Issue #2 §16 DoD)

| # | Criterion | Status | Notes |
|---|-----------|--------|-------|
| 1 | DIRECT works | **SKIPPED (already done)** | `direct_test` + `*.ru` DIRECT / PinDirectIP |
| 2 | RELAY TCP real traffic | **PARTIAL** | Path + diag (`first_byte`/`transfer_done`); needs device re-verify |
| 3 | RELAY UDP real traffic | **PARTIAL** | Code path exists; device matrix still open |
| 4 | Relay after idle continues | **DONE (code)** | Session `MaxIdleTimeout=120s`, `KeepAlive=15s` (was handshake 8–30s — root cause) |
| 5 | Transport loss → recovery | **DONE (code)** | Windows stall watchdog + `cmd=recover`; Android `ReconnectRelay` |
| 6 | Relay failover | **SKIPPED (already done)** | Catalog failover in `home_screen` / `relay_picker` |
| 7 | TUN not blackhole | **DONE (code)** | Stale `0.0.0.0/0` cleared on **stop** + start |
| 8 | Underlay not in TUN | **SKIPPED (already done)** | `protect` / `BindPhysicalUnderlay` |
| 9 | Android screen off/Doze | **PARTIAL** | Keepalive + reconnect hooks; **device evidence pending** |
| 10 | Android Wi-Fi/LTE | **DONE (code)** | `NetworkCallback` → `reconnectRelay` |
| 11 | Windows sleep/wake | **DONE (code)** | `AppLifecycleState.resumed` → `recoverAfterResume` |
| 12 | Windows network reconnect | **PARTIAL** | sing-tun monitor + stall recover; NIC-change dedicated probe pending device |
| 13 | Stale routes after crash/stop | **DONE (code)** | `ClearStaleTunnelDefaultRoute` on start **and** stop |
| 14 | DNS after idle | **DONE (code)** | `InvalidateAfterIdle` clears cache + pins |
| 15 | Split routing | **SKIPPED (already done)** | Product `split` + diagnostics modes |
| 16 | RU → DIRECT | **SKIPPED (already done)** | ADR-018 / `DefaultMode=RELAY` + PinDirect |
| 17 | Relay dest → RELAY | **SKIPPED (already done)** | must-relay; no silent DIRECT |
| 18 | No UDP/443→DIRECT workaround | **SKIPPED (already done)** | Only diagnostic `tcp_only` DROP |
| 19 | No silent RELAY→DIRECT | **SKIPPED (already done)** | `bridge.go` must-relay |
| 20 | No wide CDN CIDR | **SKIPPED (already done)** | No Cloudflare `/12` |
| 21 | Physical E2E evidence on PR | **OPEN** | Must attach logs after device runs |

---

## What was implemented this session

### Critical fix — QUIC idle (likely root cause of “internet dies after idle”)

`hyconfig.ConnectWithFallback` previously set `QUICConfig.MaxIdleTimeout` to the **handshake budget** (4–30s). Keepalives could not survive sleep/long idle.

- Now: `SessionMaxIdleTimeout = 120s`, `SessionKeepAlivePeriod = 15s`
- Handshake dial still timed by per-candidate budget separately

### Windows

- `ClearStaleTunnelDefaultRoute` on `stop()`
- Lifecycle logs: `IDLE` / `TRAFFIC_STALLED` / `RECONNECTING` / `RELAY_FAILED`
- Session stall monitor → `recoverRelay` (rebind underlay + Hysteria, TUN stays)
- IPC `cmd=recover`
- Flutter: resume → `WindowsTrafficEngine.recoverAfterResume`
- Removed false `traffic_ready` watchdog (no longer marks ready without `first_byte`)
- Softened `stream_open_no_data`: kill only when `tx==0 && rx==0` (not OR)

### Android

- `ReconnectRelay` in go_core mobile + `TunnelBridge.reconnectRelay`
- `ConnectivityManager.NetworkCallback` on Wi-Fi/LTE change → reconnect

### DNS

- `Cache.Clear`, `ClearSessionMaps`, `InvalidateAfterIdle`

### Entitlement (prior work in same branch)

- Connect blocked without active subscription / healthy relays
- `_GlassCard` Material wrap (Switch crash)

### DefaultMode SSOT (§15)

- **Not changed.** Runtime remains `DefaultMode=RELAY` + PinDirectIP (ADR-018). Documented as intentional; no unilateral flip to DIRECT.

---

## Forbidden items — compliance

| Forbidden | Compliance |
|-----------|------------|
| UDP/443 → DIRECT product | OK — not added |
| Silent RELAY → DIRECT | OK — must-relay unchanged |
| Change DefaultMode for bugfix | OK — not changed |
| Network Mode in main UX | OK — stays Diagnostics |
| Wide CIDR | OK — not added |
| Handshake = traffic proof | Fixed Windows false watchdog |

---

## Remaining for formal Issue #2 close

1. Physical Android + Windows acceptance matrix (§13) with timestamps / TX-RX / routes  
2. Attach evidence to PR  
3. Optional: rebuild `streampasscore.aar` for Android after `ReconnectRelay`  
4. Human sign-off on DefaultMode SSOT if Product wants DIRECT default again  

---

## Files touched (Issue #2)

- `client/go_core/internal/hyconfig/connect_fallback.go` (+ test)
- `client/go_core/desktop/runtime.go`, `main.go`
- `client/go_core/internal/tunbridge/bridge.go`
- `client/go_core/internal/dnscache/*`
- `client/go_core/mobile/tunnel.go`
- `client/android/.../StreamPassVpnService.kt`, `TunnelBridge.kt`
- `client/lib/services/windows_traffic_engine.dart`, `windows_core_client.dart`
- `client/lib/screens/main_shell.dart`
- (+ prior: home/servers/relay_picker entitlement + Material fix)
