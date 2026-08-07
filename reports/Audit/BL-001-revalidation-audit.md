# BL-001 Re-validation Audit (Team Lead gate)

> Дата: 2026-08-08  
> Тип: Hardening + Verification (не rewrite transport)  
> Статус: **In Progress** — code audit + unit PASS; physical Android E2E **blocked** (нет adb device)

## Вердикт

Базовый Hysteria2 transport **реализован** (`hysteria/core/v2 v2.6.1`), архитектура Flutter → VPNService → Go Core → Relay соблюдена.  
BL-001 **нельзя** закрыть Done без физического Android E2E evidence (handshake/TCP/UDP/protect/disconnect на устройстве).

Новое ADR / новая Hysteria-библиотека **не нужны**.

## Flow (подтверждено)

```
setSocketProtector → PrepareRelay (handshake до TUN)
  → VpnService.establish TUN
  → StartTunnel (Decision + tunbridge)
  → Connected
tearDown: TUN close → StopTunnel → clearProtector → disconnected
```

Ключевые файлы: `mobile/tunnel.go`, `hyconfig/*`, `tunbridge/bridge.go`, `protect/*`, `StreamPassVpnService.kt`, `TunnelBridge.kt`.

## Checklist BL-001.1–.7

| ID | Область | Статус | Примечание |
|----|---------|--------|------------|
| .1 | Audit flow | **PASS** | Prepare→Start→Stop; handshake до TUN |
| .2 | Config/TLS | **PASS*** | pinSHA256 / insecure; *pin unit tests добавлены |
| .3 | Protected underlay | **PASS** (E2E verify GAP) | protect до PrepareRelay; UDP+TCP underlay |
| .4 | TCP/UDP data path | **PASS** | Live TCP foreign IP + live UDP DNS via Hysteria |
| .5 | Fallback | **PASS** | UDP 443→8443→24443 → TCP 8443→24443 (BL-017; без TCP/443) |
| .6 | Lifecycle | **GAP*** | Close path OK; Stop 8s soft-timeout; OnDisconnected unused |
| .7 | Logging/secrets | **FIXED** | D1: URI prefix в ошибке схемы — убран |

## Routing policy (07.4)

| Проверка | Результат |
|----------|-----------|
| Transport решает DIRECT/RELAY | **Нет** — только Decision Engine |
| silent must-relay → DIRECT | **Нет** |
| UDP/443 → DIRECT | **Нет** (diag drop only) |
| DefaultMode изменён в BL-001 | **Нет** (не трогать) |

## Defects

| ID | Sev | Действие |
|----|-----|----------|
| D1 | Med | **FIXED** — `StreamPassVpnService`: не логировать `connectionConfig.take(32)` |
| D2 | Low | Open — Go `OnConnecting`/`OnDisconnected` unused (Android эмитит сам) |
| D3 | Low | Open — `StopTunnel` wait ≤8s может вернуть при живом `runTunnel` |
| D4 | Doc | Accepted — ТЗ TCP/443 vs BL-017 TCP underlay 8443/24443 |

## Architect P0 (GitHub Issue #1)

Не коммит — issue: https://github.com/Zilola5404/StreamPass/issues/1  
Зеркало: `reports/Architecture/ISSUE-1-P0-DataPlane.md`

Требует physical baseline (Этап 0) и доказательство user data plane (не только Hysteria Connected).  
Запреты совпадают с Team Lead / `07.4`.

## Automated evidence (2026-08-08)

```
cd client/go_core && go test ./...                              → PASS
cd client/go_core && go build ./...                             → PASS
.\scripts\VerifyClientConnectivity.ps1                         → 9 PASS (handshake PASS; subscription WARN INACTIVE for diag user)
STREAMPASS_RELAY_URI=… go test -run TestIntegrationHysteria    → PASS
  - handshake OK (~0.17s)
  - foreign IP via relay: 212.43.156.33 (TCP data path)
  - HTTP HEAD via relay PASS
streampasscore.aar                                             → present 28.7 MB
adb devices                                                    → empty (physical E2E blocked)
```

Новые unit: `TestParsePinSHA256ForcesSecure`, `TestParseUnsupportedScheme`, `TestNormalizeCertHash`.
D1 fix: `StreamPassVpnService.kt` — scheme-only error, без URI prefix.
## Blockers для Done

1. **Physical Android E2E** (adb device): Connect → TCP → UDP → Disconnect → Reconnect; `protect(fd)=ok`; foreign RELAY / RU DIRECT без transport hacks.
2. Live relay handshake + TCP foreign IP: `.\scripts\VerifyClientConnectivity.ps1` / `VerifyBL003.ps1 -RelayURI …`
3. AAR/APK build confirmation после любых Go Core изменений (если менялся AAR).
4. QA handoff report.

## Запреты (соблюдены)

- Нет собственного QUIC/Hysteria / sing-box / MASQUE / WireGuard
- Нет UDP/443→DIRECT / silent RELAY→DIRECT
- Нет secrets в Git
- Backend не менялся
