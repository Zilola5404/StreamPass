# Traffic block diagnosis

Generated: 2026-08-07 23:57

| Layer | Check | Status | Detail |
|-------|-------|--------|--------|
| unit | TrafficPathDiagnosis | PASS |  |
| backend | health | PASS | {"status":"ok"}  |
| relay | TCP :443 | PASS |  |
| relay | TCP :8443 | PASS |  |
| relay | TCP :24443 | PASS |  |
| device | adb | PASS |  |
| device | connect logs | PASS | 5665 chars |
| vpn | tunnel connected | PASS |  |
| vpn | DNS via TUN | PASS | 10.10.0.1 Go dnscache (HostForIP works) |
| vpn | socket protect | PASS |  |
| vpn | split tunnel | PASS |  |
| traffic | must-relay fail | PASS | count=0 |
| traffic | decision log | PASS | RELAY=5 DIRECT=0 |
| probe | live relay | WARN | no RELAY decisions captured |

## Blockers
- none