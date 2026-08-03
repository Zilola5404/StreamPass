# BL-001 — Hysteria2 Tunnel: Code Analysis

> **Task:** BL-001 Hysteria2 Tunnel  
> **Date:** 2026-08-03  
> **Author:** Senior Developer (analysis phase)  
> **Status:** Analysis complete — **awaiting confirmation before implementation**  
> **Code changes:** None (Step 1 only)

---

## 1. Executive Summary

BL-001 требует заменить **intentional stub** в `client/go_core/mobile/tunnel.go` на реальный Hysteria2 client и собрать `streampasscore.aar` для Android.

**Текущее состояние:** цепочка Flutter → Kotlin VPNService → TunnelBridge → go_core **проводка готова**, но транспорт не работает. При Connect пользователь получает ошибку (stub или ClassNotFoundException без AAR).

**Главный вывод:** реализация BL-001 затрагивает **только client-side** (`go_core` + минимальные правки Kotlin/Gradle). Backend менять не нужно. Но есть **3 блокера** и **1 критичное несоответствие API** между слоями, которые нужно решить до или в рамках первого коммита.

---

## 2. Scope BL-001 (из документации)

### Acceptance Criteria (`ai/CurrentTask.md`)

| # | Criterion | Current |
|---|-----------|---------|
| 1 | go_core tunnel.go реализует Hysteria2 client | ❌ Stub |
| 2 | gomobile build produces streampasscore.aar | ❌ Не собирался |
| 3 | AAR integrated in Android libs | ❌ Нет `libs/`, нет dependency в Gradle |
| 4 | TunnelBridge successfully loads Go core | ⚠️ Reflection готов, AAR отсутствует |
| 5 | VPN connect routes traffic through relay | ❌ |
| 6 | Foreign IP verified | ❌ |

### Explicit scope (правила задачи)

- ✅ Менять: `go_core`, сборка AAR, минимальная интеграция Android
- ❌ Не менять: backend architecture, API, unrelated modules
- ❌ Не делать в BL-001: Decision Engine (BL-005), Rule Engine (BL-006), Fallback Strategy (BL-017)

### Out of scope for BL-001 (но важно понимать)

- Per-connection DIRECT/RELAY routing (ТЗ §5–6) — отдельные задачи
- DNS Cache / DoH (ТЗ §7)
- Telemetry send on connect from client
- iOS / desktop gomobile bind

---

## 3. Data Flow (as-is)

```
HomeScreen._toggleConnection()
  → VpnChannel.connect(RelayServer)
    → MethodChannel "streampass/vpn" connect
      → MainActivity.requestConnect(args)
        → StreamPassVpnService.start(context, args)
          → establishTunnel()
            → VpnService.Builder().establish()  → TUN fd
            → TunnelBridge.startTunnel(fd, host, port, connectionConfig)
              → [reflection] streampasscore.Streampasscore.startTunnel(...)
                → go_core/mobile.StartTunnel(fd, host, port, authPassword, cb)
                  → STUB: cb.OnError("...still stubbed")
```

### Event flow back to UI

```
go_core StatusCallback
  → TunnelBridge Proxy → onState(event, relay, pingMs, error)
    → StreamPassVpnService.emit()
      → EventChannel eventSink
        → VpnChannel.statusStream
          → HomeScreen._onStatus()
```

**Вывод:** event pipeline полностью рабочий. Проблема только в go_core transport и отсутствии AAR.

---

## 4. File Inventory

| File | Role | State | BL-001 touch? |
|------|------|-------|---------------|
| `client/go_core/mobile/tunnel.go` | Go entry: StartTunnel, StopTunnel | **Stub** | ✅ Primary |
| `client/go_core/go.mod` | Module + mobile replace | Broken replace | ✅ Fix dependency |
| `client/go_core/README.md` | Build/integration guide | Documented | Maybe update |
| `client/android/.../StreamPassVpnService.kt` | TUN setup, calls bridge | Working scaffold | ⚠️ Minimal (StopTunnel) |
| `client/android/.../TunnelBridge.kt` | Reflection bridge to AAR | Working | ⚠️ Verify only |
| `client/android/.../MainActivity.kt` | MethodChannel, VPN permission | Working | ❌ No change expected |
| `client/android/app/build.gradle.kts` | Android build | **No AAR dep** | ✅ Add libs dep |
| `client/lib/services/vpn_channel.dart` | Dart → native args | Working | ❌ No change expected |
| `client/lib/services/streampass_api.dart` | GET /servers | Working | ❌ No change |
| `client/lib/screens/home_screen.dart` | Connect UI | Working | ❌ No change |
| Backend relay API | Returns connection_config | Working | ❌ No change |

### Missing artifacts

| Artifact | Expected location | Status |
|----------|-------------------|--------|
| `streampasscore.aar` | `client/android/app/libs/` | **Not found** |
| `vendor-src/mobile` | Referenced in go.mod replace | **Not found** |
| `android/app/libs/` directory | — | **Does not exist** |
| go_core unit tests | — | **None** |

---

## 5. Current Code Analysis

### 5.1 `tunnel.go` (stub)

```go
func StartTunnel(fd int, relayHost string, relayPort int, authPassword string, cb StatusCallback) {
    cb.OnConnecting()
    cb.OnError("Go core tunnel binding is present but the transport implementation is still stubbed")
}
func StopTunnel() {}
```

**Observations:**
- Callback interface matches README and TunnelBridge expectations ✅
- `StartTunnel` is **synchronous** from caller's perspective — long-running work must run in goroutine inside Go
- `StopTunnel` is empty — no lifecycle management
- No global state, no mutex — race if StartTunnel called twice
- Parameter name `authPassword` is **misleading** (see §6.1)

### 5.2 `StreamPassVpnService.kt`

**Working:**
- Receives relay data from API via Intent extras (`id`, `host`, `port`, `connectionConfig`)
- Creates TUN: `10.10.0.2/32`, DNS `1.1.1.1`/`1.0.0.1`, route `0.0.0.0/0`, MTU 1400
- Foreground service + notification
- Error handling with explicit failed state (no fake success)

**Issues for BL-001:**

| Issue | Detail |
|-------|--------|
| **StopTunnel never called** | `tearDown()` closes TUN fd but does not invoke Go `StopTunnel()` |
| **Full tunnel routing** | `addRoute("0.0.0.0", 0)` sends ALL traffic to TUN — acceptable for BL-001 MVP tunnel test, but contradicts future Decision Engine |
| **4th param semantics** | Passes `connectionConfig` (full URL string) where Go expects `authPassword` |
| **Direct import commented** | Uses reflection via TunnelBridge instead of `import streampasscore.Streampasscore` |

### 5.3 `TunnelBridge.kt`

**Design:** reflection-based loading so app compiles **without** AAR present.

**Behavior:**
- `Class.forName("streampasscore.Streampasscore")` → ClassNotFoundException if no AAR
- Maps callback methods: `onConnecting`, `onConnected`, `onDisconnected`, `onError`
- Invokes `startTunnel(Long fd, String host, Long port, String password, StatusCallback)`

**Observations:**
- Reflection adds fragility (method signature drift) but allows dev without AAR — intentional tradeoff
- No call to `stopTunnel()` anywhere in codebase
- Error messages in Russian — consistent with app

### 5.4 `VpnChannel.dart` / `home_screen.dart`

**Working correctly:**
- Passes full `RelayServer` including `connectionConfig` from `GET /api/v1/servers`
- Subscription gate before connect
- Auto-connect if settings + active subscription
- Relay selection: healthy first, sort by load_ratio then rtt_ms

**No changes needed for BL-001** unless connect verification requires telemetry hookup (out of scope).

### 5.5 Backend `connection_config`

From `backend/internal/domain/relay/relay.go`:
```go
// Server is a single relay endpoint (currently backed by Hiddify Manager:
// Xray/Reality, Hysteria2).
ConnectionConfig string
```

Example format from ТЗ / docs:
```
hysteria2://streampass-secure-auth@212.43.159.198:443/?obfs=salamander&obfs-password=streampass-relay-2024#StreamPass
```

**Implication:** go_core must parse **full hysteria2:// URI**, not just use `host` + `port` + password separately. Fields `relayHost`/`relayPort` may be redundant if URI is complete — but should remain for backward compatibility / validation.

---

## 6. Critical Issues (must address in BL-001)

### 6.1 🔴 API semantic mismatch: `authPassword` vs `connectionConfig`

| Layer | 4th parameter | Actual content passed |
|-------|---------------|----------------------|
| `tunnel.go` | `authPassword string` | — |
| `TunnelBridge.kt` | `authPassword: String` | `connectionConfig` from service |
| `StreamPassVpnService.kt` | calls with | `connectionConfig` (full hysteria2:// URL) |

**Impact:** Implementing Hysteria2 using only `authPassword` as password string will **fail**. Must parse `connectionConfig` URI or rename parameter to `connectionConfig` across Go/Kotlin (minimal rename in Go export + Kotlin reflection arg — no Dart change needed).

**Recommendation:** In `tunnel.go`, treat 4th param as `connectionConfig` (hysteria2:// URI). Use `relayHost`/`relayPort` as fallback if URI empty. **Rename in Go** for clarity; Kotlin already passes correct data.

### 6.2 🔴 `vendor-src/mobile` missing

`client/go_core/go.mod`:
```go
replace golang.org/x/mobile => ../vendor-src/mobile
```

Directory `vendor-src/mobile` **does not exist**. gomobile bind cannot run without fixing this.

**Options (need decision Q-001 adjacent):**
1. Vendor `golang.org/x/mobile` into `vendor-src/mobile` (consistent with backend approach)
2. Remove replace, use network `go get golang.org/x/mobile` (if proxy available)
3. Document manual install path

**Project rule ADR-002:** prefer vendoring when proxy unavailable.

### 6.3 🔴 No AAR build pipeline

- No `android/app/libs/` folder
- `build.gradle.kts` has **no** `implementation files('libs/streampasscore.aar')`
- README references `android/app/build.gradle` (Groovy) but project uses **Kotlin DSL** (`.kts`)

**BL-002 overlap:** AAR integration is listed as BL-002, but BL-001 acceptance criteria include "gomobile build produces AAR". Practically BL-001 and BL-002 are **one sequential work unit**.

### 6.4 🟡 `StopTunnel()` not wired

On disconnect (`tearDown()`), TUN fd is closed but Go tunnel goroutine/state not stopped. Risk: goroutine leak, incomplete Hysteria2 session teardown.

**Minimal fix:** Call `Streampasscore.stopTunnel()` from `tearDown()` via TunnelBridge (new method).

### 6.5 🟡 Open question Q-001: Hysteria2 library choice

README suggests `github.com/apernet/hysteria` (`core/client` package).

**Not verified in this analysis:**
- gomobile compatibility (cgo, unsupported packages)
- Binary size on Android
- TUN fd integration API

Alternatives mentioned in OpenQuestions: sing-box, official hy2 client.

**Recommendation:** Spike sub-step — verify `gomobile bind` succeeds with chosen library **before** full tunnel logic.

---

## 7. gomobile Constraints (implementation guardrails)

From README and gomobile rules:

| Constraint | Impact on BL-001 |
|------------|------------------|
| Only exported funcs in `package mobile` | StartTunnel, StopTunnel ✅ |
| Only primitive types + interfaces in signatures | Current signature OK ✅ |
| No Go structs in API | Must use flat params ✅ |
| `fd int` from Android | Convert via `os.NewFile(uintptr(fd), "tun")` |
| Long-running work | Must use goroutine; callbacks thread-safe |
| Generated Java package | `streampasscore` from module name `streampass/go_core` → bind path `./mobile` |

**StopTunnel** must be exported and callable from Kotlin for clean disconnect.

---

## 8. TUN Configuration Analysis

Current Android TUN setup (`StreamPassVpnService.kt`):

| Setting | Value | Notes |
|---------|-------|-------|
| Client IP | 10.10.0.2/32 | Point-to-point |
| DNS | 1.1.1.1, 1.0.0.1 | Cloudflare |
| Routes | 0.0.0.0/0 | Full tunnel |
| MTU | 1400 | Reasonable for VPN |

**For BL-001 MVP:** full tunnel is acceptable to verify foreign IP (acceptance criterion #6).

**Future (BL-005+):** Decision Engine must split DIRECT/RELAY — cannot rely on single default route forever. Not blocking BL-001.

---

## 9. Dependency & Build Environment

| Requirement | Status |
|-------------|--------|
| Go 1.22.2 | ✅ go.mod |
| gomobile CLI | TODO: not verified installed |
| Android NDK | ✅ via Flutter (`ndkVersion = flutter.ndkVersion`) |
| JDK 17 | ✅ build.gradle.kts |
| Hysteria2 Go library | ❌ Not in go.mod |
| Network/proxy for go get | Unknown — may need vendor (ADR-002) |

**Note:** Root `go.sum` absent — go_core module is separate (`streampass/go_core`).

---

## 10. Test Coverage Gap

| Area | Tests | BL-001 need |
|------|-------|-------------|
| go_core/tunnel | None | Unit test for URI parsing (if extracted) |
| TunnelBridge | None | Manual / instrumented |
| StreamPassVpnService | None | Manual on device |
| E2E VPN | None | Required for DoD — real Android device |

Per `docs/15_DefinitionOfDone.md`: new business logic should have unit tests. URI parsing helper is testable without gomobile.

---

## 11. Security Notes (no changes, awareness only)

- `connection_config` contains relay secrets (password, obfs) — passed through Intent extras in-process ✅
- Stored in PostgreSQL plaintext (known limitation S-02) — out of BL-001 scope
- Full tunnel captures all device traffic while connected — expected for VPN test

---

## 12. Proposed Implementation Plan (for confirmation — NOT executing)

### Phase A: Unblock build (prerequisite)

1. Resolve `golang.org/x/mobile` dependency (vendor-src/mobile or direct require)
2. Add Hysteria2 client library to `go_core/go.mod` (after Q-001 decision)
3. Verify `gomobile bind -target=android` compiles (even with stub logic)

### Phase B: Implement tunnel (BL-001 core)

1. Rename/clarify 4th param → `connectionConfig` in `tunnel.go`
2. Parse hysteria2:// URI (or use host/port/password fallback)
3. Open TUN from fd: `os.NewFile(uintptr(fd), "tun")`
4. Start Hysteria2 client in goroutine, bridge packets TUN ↔ HY2
5. Invoke callbacks: OnConnecting → OnConnected (with RTT) / OnError
6. Implement StopTunnel with context cancel + waitgroup

### Phase C: Android integration (BL-001 + BL-002)

1. Create `android/app/libs/`, add AAR (not committed per project rules — document build step)
2. Add to `build.gradle.kts`: `implementation(files("libs/streampasscore.aar"))`
3. Add `stopTunnel()` to TunnelBridge, call from `tearDown()`
4. Manual test: Connect → check ifconfig.me / ip

### Phase D: Verification

1. Device test with real relay (connection_config populated in DB)
2. Verify error paths: empty host, empty config, bad URI, no AAR
3. `go test` for any extracted pure Go helpers

**Estimated files to modify:**

| File | Change type |
|------|-------------|
| `client/go_core/mobile/tunnel.go` | Major — Hysteria2 impl |
| `client/go_core/go.mod` | Add deps, fix mobile |
| `client/go_core/mobile/*.go` | Possible — URI parser, client wrapper |
| `client/android/app/build.gradle.kts` | Add AAR dependency |
| `client/android/.../TunnelBridge.kt` | Add stopTunnel() |
| `client/android/.../StreamPassVpnService.kt` | Call stop on tearDown |
| `vendor-src/mobile/` | Add if vendoring |
| `docs/11_Decisions.md` | ADR for Hysteria2 lib choice (required by project rules) |

**Files NOT to modify (per rules):**
- Backend (`backend/**`)
- Flutter Dart (unless connect args change — not expected)
- Architecture layers / API

---

## 13. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Hysteria2 lib incompatible with gomobile | Medium | Blocks BL-001 | Early spike bind test |
| gomobile + NDK version mismatch on Windows | Medium | Build fails | Document exact versions from successful build |
| connection_config empty in DB for relays | Medium | Connect fails at runtime | Admin must register relay with URI |
| Full tunnel breaks local RU services | Expected | UX issue | Accept for BL-001; BL-005 fixes |
| AAR size / APK bloat | Low | UX | Monitor APK size |

---

## 14. Open Questions Requiring Confirmation Before Step 2

| ID | Question | Blocks |
|----|----------|--------|
| Q-001 | Library: `github.com/apernet/hysteria` vs sing-box vs other? | Phase B |
| NEW-Q1 | Rename Go param to `connectionConfig` — approve? | Phase B |
| NEW-Q2 | BL-001 includes Gradle AAR wiring (BL-002 overlap) — approve combined delivery? | Phase C |
| NEW-Q2 | Vendor golang.org/x/mobile into vendor-src/mobile? | Phase A |
| NEW-Q3 | Test relay: is there a live relay with populated connection_config in deployed DB? | Phase D |

---

## 15. Definition of Done Preview (BL-001)

When implementing (Step 2+), applicable from `docs/15_DefinitionOfDone.md`:

- [ ] `go build` in go_core succeeds
- [ ] gomobile produces AAR
- [ ] Android app connects on real device
- [ ] Foreign IP verified
- [ ] `StopTunnel` cleans up on disconnect
- [ ] ADR added for Hysteria2 library choice
- [ ] Update: `docs/03_CurrentState.md`, `docs/10_Progress.md`, `docs/04_Backlog.md`, `ai/LastSession.md`
- [ ] No backend changes
- [ ] No unrelated refactoring

---

## 16. Conclusion

**BL-001 is feasible without architecture changes.** The Android bridge and Flutter UI are ready. Work concentrates in `client/go_core` (~1–3 Go files) plus minimal Kotlin/Gradle wiring.

**Blockers before coding:**
1. Choose Hysteria2 Go library (Q-001)
2. Fix golang.org/x/mobile dependency
3. Resolve `authPassword` → `connectionConfig` semantic mismatch

**Recommendation:** Proceed to Step 2 with **Phase A (build unblock spike)** first — confirm gomobile + hysteria lib compiles to AAR before writing full tunnel logic.

---

> ⏸ **STOP — awaiting user confirmation to proceed to Step 2 (implementation).**
