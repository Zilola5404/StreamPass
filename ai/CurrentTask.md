# Current Task

> Updated: 2026-08-08

## Active task

**BL-001 + Architect Issue #1** — transport re-validation + data-plane stages

### Статус

**In Progress** — automated stages 1–10 largely PASS; **physical Android E2E** still blocked (no adb).

Ship candidate: **v0.1.1+35** (`routing-policy-v1`)

### Progress

| Stage | Status |
|-------|--------|
| 0 Baseline (auto) | PASS partial — device UI matrix pending |
| 1 DIRECT | PASS code/unit; **IPv6 bypass** in +35 |
| 2 RELAY TCP | PASS live + **`[vpn] traffic_ready`** |
| 3 RELAY UDP | PASS live DNS via Hysteria |
| 4 DNS/HostForIP | PASS |
| 5–7 Split/Fallback/QUIC policy | PASS (no UDP443→DIRECT / silent RELAY→DIRECT) |
| 8–10 Bypass/routes/MTU | PASS code |
| AAR + APK +35 | BUILT locally |
| Physical E2E / OTA publish | PENDING (device + deploy creds) |

### Reports

- `reports/Architecture/ISSUE-1-stage-progress.md`
- `reports/Architecture/ISSUE-1-P0-DataPlane.md`
- `reports/Audit/BL-001-revalidation-audit.md`
- Issue: https://github.com/Zilola5404/StreamPass/issues/1

### Next (auto when device appears)

1. `adb install -r …StreamPass-v0.1.1+35-signed-arm64.apk`
2. Stage 0 physical matrix (Direct Test / Split ya.ru / Full Relay / Госуслуги)
3. Expect logs: `protect(fd)=true`, `ipv6=bypass`, `[vpn] traffic_ready`
4. OTA + config `latest_client_version=0.1.1+35`
5. QA handoff → BL-001 Done
