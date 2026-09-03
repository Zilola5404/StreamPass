# RELEASE-001 — Final MVP Acceptance

**Date:** 2026-09-04  
**Branch:** `release/mvp-1`  
**RC policy:** single Release Candidate — do not test a mix of `fae42dd` and PR #3 tips.

---

## 1. Release Candidate identity

| Item | Value |
|------|--------|
| **Release Candidate Branch** | `release/mvp-1` |
| **Contains** | `fae42dd` (Architect Stage 1 / Billing baseline) **+** all Network Core fixes from PR #3 / `fix/client-network-diagnostics` **+** BILLING-001/002 **+** RELEASE-NETWORK-001 **+** Windows underlay/route harden from `main` (cherry-picked, not full merge) |
| **RC SHA (build & test)** | `2bc74bcabe7a1696270fe27bc6dc6f8f5e96aa91` (`2bc74bc`) |
| **Branch tip** | tip of `release/mvp-1` (docs may sit 1 commit above code freeze) |
| **Do not test** | random mix of `origin/main` tip alone vs PR tip alone |

### 1.1 Why not `git merge origin/main`

`origin/main` after `f160db5` diverged with Windows TUN verify commits that call a **4-arg** `NewAtomicEngineFromJSON` while decision API is **2-arg** — merge would not compile cleanly and would replace async ENGINE_STARTED / SplitRU with sync-relay-before-TUN.

**RC approach:** keep PR #3 network model (async relay, SplitRU, stall/NIC recovery) and port only safe `main` pieces:

- underlay session cache (`HasUnderlay` / skip second Get-NetRoute)
- no sing-tun `InterfaceMonitor` (hang fix)
- `route_windows.go` session gateway cleanup + stale route scan
- `InvalidateUnderlayCache` on `recoverRelay` (Wi‑Fi ↔ LTE)

### 1.2 Commit chain (must be ancestors of RC)

```
fae42dd  Architect Stage 1: Windows RU split, 3-day trial, Platega
0cea0b4  BILLING-001
b5f96bf  RELEASE-NETWORK-001
10e072d  BILLING-002
2bc74bc  RELEASE-001: RC branch + Windows protect harden from main
```

---

## 2. Acceptance matrix (Device = physical proof)

Fill **Device** on real Android / Windows before Production. Code column = present in this RC.

### Android — STEPS 2–4, 6–7

| # | Test | Expectation | Code | Device |
|---|------|-------------|------|--------|
| 2a | Launch → Login → Home | App usable with Backend ON | PASS | OPEN |
| 2b | DIRECT: ya.ru, 2ip.ru, Госуслуги, S7 | Pages load (not only Connected) | PASS | OPEN |
| 2c | RELAY known blocked dest | Request → Response → Page loaded | PASS | OPEN |
| 3a | Idle 15 min → Internet | Session alive, traffic works | PASS | OPEN |
| 3b | Idle 30 min → Internet | Same | PASS | OPEN |
| 4a | Wi‑Fi → LTE | Internet works | PASS | OPEN |
| 4b | LTE → Wi‑Fi | Internet works | PASS | OPEN |
| 6 | Connected → Relay OFF | Recovery **or** clean disconnect + ISP Internet | PASS | OPEN |
| 7 | Backend OFF → Connect | «Сервер временно недоступен…» / no infinite spinner; no blackhole | PASS | OPEN |

### Windows — STEPS 5–7

| # | Test | Expectation | Code | Device |
|---|------|-------------|------|--------|
| 5a | Connect → Internet | Page load | PASS | OPEN |
| 5b | Idle 30 min → Internet | Works | PASS | OPEN |
| 5c | Sleep → Wake → Internet | recover path | PASS | OPEN |
| 5d | Disconnect → normal Internet | No stale `0.0.0.0/0` via StreamPass | PASS | OPEN |
| 6 | Relay OFF while Connected | Recovery or clean disconnect | PASS | OPEN |
| 7 | Backend OFF → Connect | User-facing unavailable; no hang | PASS | OPEN |

### STEP 8 — Real Payment

| # | Test | Expectation | Code | Device |
|---|------|-------------|------|--------|
| 8 | Trial expired → Paywall → Basic → Platega → Webhook → ACTIVE → Connect | Full money path | PASS | OPEN |

---

## 3. Gate rules (organizational)

1. **One RC SHA** is the only artifact for QA / beta.  
2. Saying «всё запушено» is not acceptance — cite **RC SHA**.  
3. PR #3 merge to `main` only after Device columns for STEPS 2–7 are PASS (STEP 8 may be staged beta).  
4. Public launch remains **blocked** until Device matrix + real Platega are green.  
5. Closed beta: allowed after RC freeze + Android/Windows smoke (STEPS 2, 5, 7 minimum).

---

## 4. Rebuild before device tests

1. Android AAR: `go_core` with `-tags with_gvisor` matching this RC.  
2. Windows: rebuild `streampasscore.exe` from `client/go_core/desktop` on this SHA.  
3. Attach logcat / connect.log / core logs to QA notes per failed row.

---

## 5. Verdict (pre-device)

| Area | Status |
|------|--------|
| Billing (`fae42dd` + BILLING-001/002) | Accept — do not redesign trial/plans/Platega |
| Network code (PR #3 + RELEASE-NETWORK-001 + Windows protect harden) | Accept for RC code freeze |
| Proven on devices | OPEN — ~60% until matrix filled |
| Closed beta readiness | After Device smoke + RC push |
| Public launch | Not yet |

**Next human action:** run STEPS 2–8 against **this RC SHA only**, mark Device column, then Code Review → merge PR #3 → Production cut from `release/mvp-1`.
