# Issue #5 / RELEASE-001 — Windows + Lab Evidence (2026-09-10)

**Issue:** https://github.com/Zilola5404/StreamPass/issues/5  
**RC tip:** `8dc97dc` (+ this evidence commit)  
**Branch:** `release/mvp-1`  
**Host:** Windows (non-Admin agent session)  
**Android adb:** empty — physical Android deferred until device USB/wireless debugging attached

---

## 1. Audit vs Issue #5 (reuse, no second architecture)

| Issue #5 area | Existing component used | Status |
|---------------|-------------------------|--------|
| Decision / Rule Engine | `go_core/internal/decision` | Lab PASS |
| DIRECT/RELAY SSOT | `DefaultMode=DIRECT` (ADR-020) | Docs+code aligned |
| TRAFFIC_READY ≠ CONNECTED | `ConnectionController` + `markTrafficReady` | Lab PASS |
| Reconnect latch | `SetHysteriaClient` resets latch | Code in `8dc97dc` |
| Server unavailable / no blackhole | Issue #4 UX + Windows tear-down | Lab PASS (unit + ISP intact) |
| Billing / trial / Platega | BILLING-001/002 | Unit PASS; live Platega OPEN |
| Forbidden workarounds | No UDP/443→DIRECT product path | Confirmed absent |

---

## 2. Routing matrix (Issue #5 destinations) — Lab

Command: `go test ./internal/decision -run 'TrafficMatrix|Issue5_S7andGithub|DecideDetailed' -count=1`

| Destination | Expected | Result |
|-------------|----------|--------|
| ya.ru / yandex | DIRECT | PASS |
| 2ip.ru | DIRECT | PASS |
| gosuslugi.ru | DIRECT | PASS |
| s7.ru / www.s7.ru | DIRECT (`*.ru`) | PASS (`TestIssue5_S7andGithub`) |
| github.com | RELAY | PASS |
| youtube.com | RELAY | PASS |
| openai.com | RELAY | PASS |
| unknown (example.com) | DIRECT | PASS |
| cloudflare.com | DIRECT (unknown) | PASS |

---

## 3. Windows physical / host checks (this machine)

| Test | Result | Evidence |
|------|--------|----------|
| VerifyWindowsTUN stages 5–10 + IPC ping | **PASS** | `reports/QA/TASK-WIN-001-stages5-12-2026-09-10.md` |
| Rebuild `streampasscore.exe` (`with_gvisor`) from tip | **PASS** | native exe ~18.3 MB, 2026-09-10 16:10 |
| Baseline route table: no `10.10.0.x` / StreamPass default | **PASS** | `route print -4` — STALE_NONE |
| ISP public IP before/after failed elevated TUN attempt | **PASS** | `80.80.117.213` unchanged |
| Non-Admin `cmd=start` (direct_test) | **PARTIAL** | CONNECT_START + underlay bind; CreateAdapter needs Admin — **no blackhole**, routes clean after kill |
| Live Wintun CreateAdapter / CONNECT / page load | **SKIP** | Agent session `Admin=False` — run elevated: `scripts\VerifyWindowsTUN.ps1` |
| Sleep/wake, idle 30m, full Flutter UI Connect | **OPEN** | Needs interactive Admin + app session |
| Real RELAY handshake + foreign page load | **OPEN** | Needs backend/relay + Admin TUN |

---

## 4. Flutter / UX (Issue #5 §§10–11)

| Test | Result |
|------|--------|
| `traffic_behavior_test` + `connection_controller_test` | PASS |
| `user_facing_errors_test` (Сервер временно недоступен…) | PASS |
| `e2e_flow_test` Login/Register → Home | PASS |

---

## 5. Android physical (Issue #5 §17)

| Item | Status |
|------|--------|
| `adb devices` | **empty** |
| Install / Connect / DIRECT / RELAY / idle / Wi‑Fi↔LTE | **BLOCKED** |

**Handoff:** attach phone with USB debugging (or `adb connect <ip>:5555`), confirm `adb devices` shows `device`, then re-run agent / provide wireless ADB.

---

## 6. Gate for Issue #5 Acceptance

| Criterion | Status |
|-----------|--------|
| Lab routing + traffic_ready + no forbidden workarounds | PASS |
| Windows contract VerifyWindowsTUN | PASS |
| Windows live TUN + real traffic | OPEN (need Admin) |
| Android physical E2E | OPEN (need adb) |
| Live Platega | OPEN |
| **Issue #5 closable as Done** | **No** — Device evidence incomplete |

---

## 7. Operator next commands

```powershell
# Elevated PowerShell (Admin):
cd C:\01_Projects\StreamPass
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\VerifyWindowsTUN.ps1

# Android:
adb devices
# then agent continues Device matrix
```
