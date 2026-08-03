# Next Task

> After: BL-001, BL-002 (completed 2026-08-03)

## BL-003: End-to-end VPN test on Android device

### Description
Проверить полный цикл на реальном Android-устройстве: Connect → TUN → Hysteria2 → relay → foreign IP.

### Steps
1. Убедиться, что relay в БД имеет заполненный `connection_config` (hysteria2:// URI)
2. Собрать и установить APK (`flutter run` или `assembleDebug`)
3. Залогиниться, выбрать relay, нажать Connect
4. Проверить статус «Подключено» в UI
5. Проверить IP через браузер (ifconfig.me / 2ip.ru) — должен быть IP relay
6. Disconnect — корректный tearDown, `StopTunnel()` вызван

### Depends On
- BL-001, BL-002 — done
- Live relay с populated `connection_config`

### After BL-003

| Order | ID | Task |
|-------|-----|------|
| 1 | BL-005 | Decision Engine on client |
| 2 | BL-006 | Rule Engine on client |
| 3 | BL-004 | Live-test ЮKassa |
| 4 | BL-010 | CI/CD GitHub Actions |
| 5 | BL-012 | Update README |
