# Windows Client — Implementation Plan (TASK-WIN-001)

> Date: 2026-08-15  
> Gate: `reports/QA/TASK-WIN-001-relay-gate-2026-08-14.md` **PASS**  
> Stages 5–12: `scripts/VerifyWindowsTUN.ps1` + `reports/QA/TASK-WIN-001-stages5-12-*.md`

## Order (TZ §28)

1. ~~Independent Relay~~ **DONE**
2. ~~Windows skeleton~~ **DONE**
3. ~~Auth + API~~ **DONE** (same `/api/v1`)
4. ~~ConnectionController~~ **DONE** (Connected ≠ handshake)
5. ~~TUN / Wintun~~ **DONE**
6. ~~DIRECT path~~ **DONE** (code + unit; live Admin E2E via Verify script / Direct Test)
7. ~~RELAY + traffic_ready~~ **DONE** (UI gate + chip «Ожидание traffic_ready…»)
8. ~~DNS + HostForIP~~ **DONE** (`198.18.0.1`)
9. ~~Decision / split~~ **DONE** (same engine; matrix tests)
10. ~~IPv6 Variant B~~ **DONE** (no Inet6 capture)
11. ~~MTU 1280/1350/1400~~ **DONE** (Diagnostics → IPC)
12. ~~Adaptive fallback~~ **DONE** (shared `ModeFallback` / blackhole / underlay)
13. ~~UI polish / telemetry hooks~~ **DONE** (desktop layout, diag rows, DiagUploader, verify script)

## Remaining before production ship

- Physical Windows E2E per `reports/Audit/AUDIT-WIN-001-Windows-Client.md` §12 (`scripts/VerifyWindowsE2E.ps1`)
- Live Admin run of `VerifyWindowsTUN.ps1` (CreateAdapter) on target PC
- Relay URIs: **`pinSHA256` required** — `insecure=1` blocked unless `STREAMPASS_ALLOW_INSECURE=1` (dev)
- Windows installer / OTA (not in MVP scope — no APK path)

## Architecture locks

```text
Unknown → DIRECT rules for *.ru; DefaultMode=RELAY for foreign (ADR-018)
DIRECT works if Relay is down
Hysteria underlay → OS default-route NIC (never TUN/Wintun)
insecure=1 forbidden in production (hard fail; dev: STREAMPASS_ALLOW_INSECURE=1)
Session-owned route cleanup (not only fixed 10.10.0.2)
No Cloudflare CIDR → RELAY
IPv6 MVP: Variant B
Platform: Wintun sidecar (ADR-019); WFP deferred (ADR-020)
```
