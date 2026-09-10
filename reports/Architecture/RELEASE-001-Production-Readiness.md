# RELEASE-001 — Baseline + Production Gate Package

**Date:** 2026-09-10  
**Branch:** `release/mvp-1`  
**Baseline tip (before this push):** `24af13f`  
**Code freeze ancestry:** `2bc74bc` contains Network+Billing harden; subsequent commits are lab + this readiness package.

---

## 1. Baseline (pre-device)

| Item | Result |
|------|--------|
| Android physical adb | **No device attached** — Device matrix remains OPEN |
| Windows physical interactive TUN | Not executed in this runner (Admin + UI session required) |
| Lab (prior): go decision/protect/tunbridge/mobile | PASS @ `24af13f` |
| Lab (prior): flutter test 103 | PASS @ `24af13f` |
| Public API health from prior runner | FAIL (unreachable) — infra env, not RC compile |

**Honest baseline:** CONNECTED≠proven on devices for this RC. Do **not** treat historical 2026-08 QA as proof for `2bc74bc`/`24af13f`.

---

## 2. Architecture SSOT (this package)

```
Destination → Classification → Rule Engine → Decision Engine → DIRECT | RELAY
Unknown → DIRECT (decision.DefaultMode = ModeDirect)
```

- Forbidden workarounds **not** used: UDP/443→DIRECT product path; silent Relay→DIRECT-all; Cloudflare /12→DIRECT.
- ADR-020 documents supersession of ADR-018 DefaultMode=RELAY for MVP.

---

## 3. Code fixes in this readiness package

1. **traffic_ready latch reset** on `SetHysteriaClient` (reconnect / async relay attach).  
2. Android underlay change: emit `connecting` → reconnect → `connected` without trafficReady until first_byte.  
3. Canonical `[lifecycle]` tokens: `CONNECT_START`, `RELAY_CONNECTING`, `RELAY_CONNECTED`, `TRAFFIC_READY`, `RECONNECT_START/SUCCESS/FAILED`.  
4. Diagnostics IPv6: Windows Variant B vs Android capture+drop+AAAA-suppress.  
5. Docs SSOT: `07.4`, `18`, `37`, ADR-020.

---

## 4. Physical E2E evidence (required for 100/100)

| Platform | Test | Result | Evidence |
|----------|------|--------|----------|
| Android | DIRECT (ya.ru, 2ip, S7, gosuslugi) | **OPEN** | — |
| Android | RELAY TCP | **OPEN** | — |
| Android | RELAY UDP | **OPEN** | — |
| Android | Split RU / foreign | **OPEN** | — |
| Android | idle 5/15/30, screen OFF/ON, WiFi↔LTE | **OPEN** | — |
| Windows | DIRECT / RELAY / Split | **OPEN** | — |
| Windows | idle / sleep-wake / stale route / disconnect | **OPEN** | — |
| Billing | Real Platega webhook → ACTIVE | **OPEN** | — |
| Negative | Backend OFF / Relay OFF / no blackhole | **OPEN** | — |

Fill this table on devices built from the tip of `release/mvp-1` after this push; attach logcat / connect.log / `route print` screenshots to the PR.

---

## 6. Quality commands (this package)

```
cd client/go_core && go test ./internal/decision ./internal/tunbridge ./internal/protect ./mobile -count=1
cd client/go_core && go vet ./internal/tunbridge ./desktop && go build -o NUL ./desktop
cd backend && go test ./internal/application/billing ./internal/domain/subscription ./internal/application/auth -count=1
cd client && flutter test
cd client && flutter analyze --no-fatal-infos
```

Notes: `flutter analyze` still reports pre-existing style `info` (e.g. `withOpacity`); warnings from unused imports cleaned in this package. Full `go test ./...` / `go vet ./...` at repo roots may include unrelated packages — scoped packages above are the RELEASE-001 network/billing surface.

---

## 5. Gate verdict (developer)

| Criterion | Status |
|-----------|--------|
| P0 code gaps from lab audit (latch, SSOT docs, lifecycle names) | Addressed in this package |
| go/flutter quality commands | See PR / commit notes |
| Physical Android E2E | **NOT PASS** |
| Physical Windows E2E | **NOT PASS** |
| Billing live Platega | **NOT PASS** |
| **100/100 MVP / Public Beta** | **NOT DECLARED** |

**Remaining known issues (non-exhaustive):** Device matrix OPEN; AAR rebuild needs JDK/`javac` on build host; live API reachability from CI runner; Google/Meta CIDR safety nets remain narrow-ish anycast nets (domain-first still required).

**Next:** Human Device E2E on this tip → attach evidence → only then merge to Production.
