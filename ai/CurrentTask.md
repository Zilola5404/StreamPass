# Current Task

> Updated: 2026-08-03

## Главная задача

**BL-003: End-to-end VPN на Android**

## Описание

Проверить на Android-устройстве полный цикл: Connect → TUN → Hysteria2 → relay → foreign IP.

## Контекст

- BL-001/BL-002 завершены: go_core transport, AAR, Gradle, StopTunnel
- Relay `connection_config` приходит из GET /servers
- Пример URI: `hysteria2://streampass-secure-auth@212.43.159.198:443/?obfs=salamander&obfs-password=streampass-relay-2024#StreamPass`

## Acceptance Criteria

- [x] go_core tunnel.go реализует Hysteria2 client
- [x] gomobile build produces streampasscore.aar
- [x] AAR integrated in Android libs
- [x] TunnelBridge successfully loads Go core
- [ ] VPN connect on Android device routes traffic through relay
- [ ] Foreign IP verified

## Files to Test

- `client/android/app/` (APK на устройстве)
- Backend relay с populated `connection_config`

## Previous Task (Completed)

BL-001 Hysteria2 tunnel + BL-002 AAR integration — completed 2026-08-03.
