# RELEASE-NETWORK-001 — Final Network Core Stabilization

**Date:** 2026-09-03  
**Priority:** P0 — BLOCKER BEFORE PUBLIC LAUNCH  
**HEAD baseline:** post-BILLING-001 (`0cea0b4`) + this commit  
**Scope:** Fix real Network Core gaps only. No redesign, no Billing change, no UI redesign.

---

## 1. Test Matrix (code vs device)

Fill **Device** column on physical Android / Windows before declaring MVP network-stable.

### Android

| Test | Code | Device |
|------|------|--------|
| DIRECT | PASS | OPEN |
| RELAY TCP | PASS | OPEN |
| RELAY UDP | PASS | OPEN |
| Split Routing | PASS | OPEN |
| Idle 5 min | PASS (keepalive + stall watchdog) | OPEN |
| Idle 15 min | PASS (code) | OPEN |
| Idle 30 min | PASS (code) | OPEN |
| Screen OFF/ON | PARTIAL (FG service + QUIC KA) | OPEN |
| Wi-Fi → LTE | PASS (NetworkCallback → ReconnectRelay) | OPEN |
| LTE → Wi-Fi | PASS | OPEN |
| Disconnect | PASS | OPEN |
| Traffic stalled → recover | PASS (new Android watchdog) | OPEN |
| traffic_ready before UI Connected | PASS (new) | OPEN |
| Failed connect / no blackhole | PASS | OPEN |

### Windows

| Test | Code | Device |
|------|------|--------|
| DIRECT | PASS | OPEN |
| RELAY TCP | PASS | OPEN |
| RELAY UDP | PASS | OPEN |
| Idle 30 min | PASS | OPEN |
| Sleep/Wake | PASS (`recover` on resume) | OPEN |
| Network reconnect (NIC change) | PASS (new underlay poll → recoverRelay) | OPEN |
| Disconnect removes `0.0.0.0/0` | PASS | OPEN |
| App crash → stale route clear on next start | PASS (best-effort Admin) | OPEN |

### How to run device matrix

1. Build APK with AAR `-tags with_gvisor` after go_core changes.  
2. Build Windows `streampasscore.exe` matching desktop tags.  
3. For each row: Connect → exercise scenario → confirm **page loads** (not only Connected).  
4. DIRECT: ya.ru, 2ip.ru, gosuslugi, S7 — must work without Relay.  
5. RELAY: known blocked destinations via Rule Engine — TCP and UDP separately.  
6. Idle: leave session quiet 5/15/30 min → open site; Connected=true & Traffic=false = **FAIL**.  
7. Blackhole: Connect → kill Relay path → expect recovery **or** clean disconnect + ISP Internet restored.  
8. Attach logcat / connect.log / Windows core logs to QA folder.

---

## 2. Code fixes in this release

| Fix | Where |
|-----|--------|
| Android `TRAFFIC_STALLED` → `ReconnectRelay` (max 2) then clean error disconnect | `go_core/mobile/tunnel.go` |
| Android `traffic_ready` → Flutter; UI Connected only after first_byte | VpnService + `connection_controller` + `vpn_channel` |
| Android traffic_ready 35s timeout → tear down | `home_screen.dart` |
| Windows NIC / underlay index change → `recoverRelay` | `go_core/desktop/runtime.go` |
| Split: **DefaultMode = DIRECT** (unknown → ISP; known blocked → RELAY rules) | `decision/rule.go` |
| Refund + support legal copy (Gate 8) | `legal_docs.dart` + onboarding |

**Not added:** permanent UDP/443→DIRECT workaround (still diagnostic-only / DROP).

---

## 3. P0-2 Server Unavailable UX

Already code-PASS (Issue #4). Reconfirm on device:

- App always reaches Home with Backend OFF.  
- Connect → «Сервер недоступен».  
- TUN not started; user Internet intact.

---

## 4. P0-3 Real payment (Gate 5)

BILLING-001 architecture is in place. **Live E2E payment remains OPEN:**

Register → Trial 72h → expire → Paywall → Order → Payment → Webhook → ACTIVE → Connect  
+ duplicate webhook → one subscription  
+ canceled payment → not ACTIVE  

Requires production-like Platega/YooKassa credentials — not closed by unit tests.

---

## 5. Production checklist (Gate 6) — MVP, no K8s

### Backend VPS
- [ ] Docker `restart: unless-stopped`
- [ ] Production `.env` (not in git)
- [ ] Postgres backup schedule
- [ ] HTTPS (reverse proxy)
- [ ] `GET /health` monitored
- [ ] Log retention

### Relay
- [ ] Relay A live (TCP + UDP verified)
- [ ] Relay B live (failover)
- [ ] External IP / hysteria config verified

---

## 6. Security checklist (Gate 7) — minimum

### Secrets
- [ ] No production secrets in GitHub history / `.env` committed
- [ ] Rotate any leaked payment / JWT / Telegram / DB passwords

### Webhook
- [ ] Auth headers / webhook secret required
- [ ] Status re-fetched from provider (body not trusted alone)
- [ ] Idempotent PENDING→SUCCEEDED
- [ ] Canceled status does not activate

### API
- [ ] JWT on protected routes
- [ ] User cannot grant self ACTIVE
- [ ] User cannot read others’ subscription/orders

---

## 7. Legal (Gate 8)

| Doc | Status |
|-----|--------|
| Privacy Policy | In-app |
| Terms of Service | In-app |
| Subscription / Refund | In-app (new) |
| Support contacts | `support@streampass.app` in-app |

Publish on public site before mass ads.

---

## 8. Business tariff

Keep `business` in Plan Catalog. **Do not market as full org platform** until multi-user org is ready. Prefer «Для бизнеса — скоро» on landing.

---

## 9. Explicitly deferred (per architect)

Kubernetes, microservices, AI routing, 10 relays, referral, enterprise RBAC, complex analytics, Business dashboard.

---

## 10. Definition of Done status

| Criterion | Status |
|-----------|--------|
| Code fixes for stall / traffic_ready / NIC recover | **DONE** |
| Unit tests connection policy | **DONE** |
| Physical Test Matrix filled PASS | **OPEN — required for gate close** |
| Real payment E2E | **OPEN** |
| Prod VPS + Relay A/B | **OPEN (ops)** |
| Secrets audit | **OPEN (ops)** |
| Legal hosted URLs | **PARTIAL (in-app OK)** |

**Engineering close for code path:** this commit.  
**Product close for public launch:** complete Device column + Gates 5–7.
