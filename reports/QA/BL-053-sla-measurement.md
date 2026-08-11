# BL-053 — Device SLA Measurement (Formal Pass)

> **Date:** 2026-08-10  
> **Targets (TZ section 22):** cold start <= 2s, connect <= 5s, recover <= 10s  
> **Harness:** `scripts/MeasureDeviceSLA.ps1`, `client/lib/services/sla_targets.dart`, `client/test/sla_targets_test.dart`

---

## Automated verification (2026-08-10)

| Check | Result | Evidence |
|-------|--------|----------|
| `SlaTargets` unit tests | **PASS** | `flutter test test/sla_targets_test.dart` — 3/3 |
| API `/health` <= 2s | **PASS** | `MeasureDeviceSLA.ps1 -SkipAdb` |
| API `/config` <= 2s | **PASS** | `MeasureDeviceSLA.ps1 -SkipAdb` |
| Backend auth/security tests | **PASS** | `go test ./internal/application/auth/ ./internal/infrastructure/security/ ./internal/infrastructure/http/middleware/` |

---

## Device evidence (+42, One UI 8.5, split, pl-warsaw-1:24443)

Source: user diagnostics export 2026-08-09.

| Metric | Measured | Budget | Result |
|--------|----------|--------|--------|
| Hysteria handshake (`pingMs`) | **0.746 s** | 5 s | **PASS** |
| Connect `connecting` -> `connected` (Flutter) | **5.34 s** | 5 s | **FAIL** (+340 ms) |
| Connect `session begin` -> `tunnel connected` (native) | **5.34 s** | 5 s | **FAIL** (+340 ms) |
| Recover (airplane toggle) | not measured | 10 s | **PENDING** manual |

### Breakdown (native log)

| Phase | Duration |
|-------|----------|
| PrepareRelay / hysteria ok | ~0.75 s |
| TUN establish (split exclude-ru 11406 routes) | ~4.5 s |
| **Total to Connected** | **~5.34 s** |

**Root cause of connect SLA miss:** slow TUN setup on Samsung One UI with large RU exclude list, not relay latency.

---

## Verdict

| Stage | Status |
|-------|--------|
| BL-053 harness (script + unit + API) | **CLOSED** |
| Connect <= 5s on device | **PARTIAL** — relay path OK; TUN setup exceeds budget on One UI |
| Recover <= 10s | **OPEN** — needs one manual airplane-mode test |

### Recommendations

1. Use **nl-native-1:443** for SLA retest (same as Hiddify baseline).
2. Measure recover: airplane 3s -> Connected; log `failover` / `recover` lines.
3. Future: cache split routes or lazy exclude to cut TUN time (product perf, not BL-053 blocker for harness).

---

## How to re-run

```powershell
cd C:\01_Projects\StreamPass\client
flutter test test/sla_targets_test.dart

powershell -File ..\scripts\MeasureDeviceSLA.ps1 -SkipAdb
powershell -File ..\scripts\MeasureDeviceSLA.ps1 -LogFile C:\path\to\diagnostics-export.txt
```
