# Issue #1 / BL-001 — Stage progress

> Дата: 2026-08-08  
> Sources: Team Lead BL-001 + [Architect Issue #1](https://github.com/Zilola5404/StreamPass/issues/1)  
> Device: **adb empty** — physical matrix Stage 0 partial; automated + code gates advanced

## Этап 0 — Baseline

| Check | Result | Evidence |
|-------|--------|----------|
| adb device | **BLOCKED** | `adb devices` empty after restart |
| Prior connect log | PARTIAL | `reports/QA/connect-logcat-0058.txt`: `protect=true`, DNS-in-TUN, Meta CIDR RELAY |
| Prior DiagnoseTrafficBlock | PARTIAL | retest2: unit/backend/relay PASS; device was PASS earlier session |
| Live API + Hysteria handshake | **PASS** | `VerifyClientConnectivity.ps1` |
| Live TCP foreign IP | **PASS** | `212.43.156.33` via relay |
| Decision matrix ya.ru/2ip DIRECT | **PASS** | `go test` decision package |
| quic_direct_bypass / silent RELAY→DIRECT | **ABSENT** | code review tunbridge |

**Этап 0 closed for automation; physical UI matrix deferred until USB device.**

## Этап 1 — DIRECT data path

| Check | Result |
|-------|--------|
| Decision `*.ru` / DefaultMode=DIRECT | PASS (unit) |
| `pipeTCP` DIRECT uses `protect.Control` | PASS (code) |
| Split RU excludeRoute (API 33+) | PASS (code) |
| **IPv6 blackhole risk** | **FIXED in +35**: `allowFamily(AF_INET6)` — VPN IPv4-only |
| Physical Direct Test ya.ru/2ip | PENDING device |

## Этап 2 — RELAY TCP

| Check | Result |
|-------|--------|
| Live TCP ifconfig.me via Hysteria | PASS |
| Blackhole 3s no silent DIRECT | PASS (code) |
| **`[vpn] traffic_ready`** | **ADDED** — first remote byte (handshake ≠ traffic) |
| Physical Full Relay HTTPS | PENDING device |

## Этап 3 — RELAY UDP

| Check | Result |
|-------|--------|
| `hy.UDP` Send/Receive path | PASS (code) |
| Live UDP DNS via relay | RUN in integration |
| No product UDP/443→DIRECT | PASS |
| Physical QUIC | PENDING device |

## Этап 4 — DNS / HostForIP

| Check | Result |
|-------|--------|
| VPN DNS `10.10.0.1` | PASS (code + prior logs) |
| HostForIP reverse map | PASS (unit) |
| Physical HostForIP on CDN | PENDING device |

## Этап 5–7 — Split / Fallback / QUIC

| Check | Result |
|-------|--------|
| DefaultMode=DIRECT | PASS |
| FALLBACK only ModeFallback | PASS |
| Network Mode only E09 diag | PASS (prior TASK-02) |
| Google/Meta CIDR = IP-only safety net (not Cloudflare/12) | PASS (policy) |

## Этап 8–10 — Bypass / routes / MTU

| Check | Result |
|-------|--------|
| App bypass package list | PASS (code) |
| PrepareRelay before TUN | PASS |
| protect before Prepare | PASS |
| MTU option 1200–1500 | PASS (code) |
| Physical bypass Госуслуги | PENDING device |

## Ship

- Client **v0.1.1+35**: IPv6 bypass + `traffic_ready` + D1 log fix + pin/UDP unit+integration
- AAR rebuilt 2026-08-08; APK `StreamPass-v0.1.1+35-signed-arm64.apk`
- Live: handshake PASS, TCP foreign IP PASS, **UDP DNS via relay PASS**
- Physical Stage 0 UI matrix: **still needs adb device**

## Automated evidence (2026-08-08 late)

```
go test ./...                                              PASS
TestIntegrationHysteriaUDPEcho                             PASS (61 bytes from 1.1.1.1:53)
gomobile bind → streampasscore.aar                         OK (~30 MB)
flutter build apk --release --android-arm64                OK → +35
```
