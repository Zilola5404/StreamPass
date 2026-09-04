# RELEASE-001 — Staged Lab Execution Log

**Date:** 2026-09-04  
**RC branch:** `release/mvp-1`  
**Code freeze SHA:** `2bc74bc`  
**Branch tip at execution:** see git after this commit  

**Constraint:** No Android device attached (`adb devices` empty). Public API `212-43-156-33.nip.io` unreachable from this runner. Real Platega payment and physical Idle/Sleep/Network Change **cannot** be marked Device=PASS from this environment without inventing results.

---

## STEP 1 — Single Release Candidate

| Check | Result |
|-------|--------|
| Branch `release/mvp-1` | PASS |
| `fae42dd` is ancestor | PASS |
| `2bc74bc` is ancestor | PASS |
| PR #3 tip synced (`fix/client-network-diagnostics`) | PASS (same tip) |
| Full merge `origin/main` | SKIPPED (known broken 4-arg AtomicEngine) — safe cherry-picks already in RC |

---

## STEP 2 — Android DIRECT / RELAY (Lab)

| Check | Result | Notes |
|-------|--------|-------|
| `go test ./internal/decision ./mobile …` | **PASS** | DIRECT default + RELAY matrix |
| Flutter `traffic_behavior_test` (ya.ru/2ip/gosuslugi → DIRECT; youtube/instagram → RELAY) | **PASS** | Rule contract only |
| Flutter `e2e_flow_test` Login→Home | **PASS** | After DiagUploader timer leak fix |
| Physical page load ya.ru / 2ip / Госуслуги / S7 | **BLOCKED** | No device |
| Physical RELAY page load | **BLOCKED** | No device |
| Rebuild AAR `-tags with_gvisor` | **BLOCKED** | `javac` not on PATH / no JDK install; existing `streampasscore.aar` kept |

---

## STEP 3 — Idle 15 / 30 min

| Check | Result |
|-------|--------|
| Code: stall watchdog + QUIC keepalive + IDLE grace | **PASS** (static + unit coverage in go_core) |
| Device idle 15/30 min then page load | **BLOCKED** (no device) |

---

## STEP 4 — Wi‑Fi ↔ LTE

| Check | Result |
|-------|--------|
| Code: Android `ReconnectRelay` + Windows `UNDERLAY_CHANGED` → `recoverRelay` + `InvalidateUnderlayCache` | **PASS** (present in RC) |
| Device network switch | **BLOCKED** |

---

## STEP 5 — Windows

| Check | Result |
|-------|--------|
| `go test ./internal/protect ./internal/tunbridge` | **PASS** |
| Rebuild `streampasscore.exe` (`build_core.ps1`, `with_gvisor`) | **PASS** (local binary; gitignored) |
| Connect / Idle 30 / Sleep-Wake / Disconnect stale route on device | **BLOCKED** (needs Admin + interactive session) |

---

## STEP 6 — Relay failure

| Check | Result |
|-------|--------|
| Code: relay handshake fail → `relay_unavailable` + `r.stop()` (no blackhole) | **PASS** (desktop `runtime.go`) |
| Live Relay OFF while Connected | **BLOCKED** |

---

## STEP 7 — Server OFF

| Check | Result |
|-------|--------|
| Copy: `Сервер временно недоступен. Попробуйте позже.` | **PASS** (`UserFacingErrors` + unit test) |
| Backend unit/integration (billing/auth/router) | **PASS** |
| Live Backend OFF app launch | **BLOCKED** (no device; API also unreachable here) |
| Public `/health` from runner | **FAIL** (connection refused / unreachable — env, not RC regression) |

---

## STEP 8 — Real Payment

| Check | Result |
|-------|--------|
| `go test ./internal/application/billing` + subscription domain | **PASS** (incl. webhook idempotency) |
| Real Platega money path | **BLOCKED** (requires live provider + trial-expired account) |

---

## STEP 9 — Merge / Production gate

| Action | Result |
|--------|--------|
| Merge PR #3 → `main` | **HOLD** — Device columns still OPEN |
| Production cut | **HOLD** — same |
| Push RC + lab report | **DONE** (this commit) |

**Gate reminder:** Lab PASS ≠ Device PASS. Closed beta only after physical STEPS 2+5+7 smoke on builds from this RC.

---

## Fix applied during lab (safe)

- `DiagUploader`: cancelable initial 3s flush timer so Home dispose / Flutter widget tests do not leak pending timers (`e2e_flow_test` was red).

---

## Operator checklist (human, same RC SHA only)

1. Install JDK → rebuild AAR with `gomobile bind -tags with_gvisor` → copy to `android/app/libs/`.  
2. Use rebuilt `client/windows/native/streampasscore.exe`.  
3. Fill Device column in `RELEASE-001-Final-MVP-Acceptance.md`.  
4. Only then Code Review → merge → Production.
