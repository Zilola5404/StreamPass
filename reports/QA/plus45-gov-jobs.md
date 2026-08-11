# +45 — gov bypass + jobs RELAY

> 2026-08-10 · `0.1.1+45` (versionCode 2045) · installed via adb

## Causes

| Symptom | Root cause |
|---------|------------|
| Госуслуги / ФНС / S7 «VPN включён» | `appBypass=1` — package visibility (API 30+) blocked `getPackageInfo` → `addDisallowedApplication` never ran for `ru.fns.lkfl` / `ru.s7tl.app` / `ru.rostel` |
| Indeed / Upwork / LinkedIn «not available in your country» | DNS→RELAY, TCP→`default_direct` on Cloudflare IP (empty/wrong HostForIP / anycast) → RU exit IP |

## Fixes
1. Manifest `<queries>` (known packages + LAUNCHER) + launcher heuristic in `VpnBypassApps`
2. DNS-time `PinRelayIP` + `HostsForIP` set + `/24` neighbor for Cloudflare anycast

## User
Reconnect VPN once after install. Then open ФНС/S7/Госуслуги and Indeed/Upwork/LinkedIn.
