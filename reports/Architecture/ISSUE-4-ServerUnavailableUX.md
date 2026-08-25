# Issue #4 — Server Unavailable UX / Startup & Connect Flow

> Дата: 2026-08-25  
> Ветка: `fix/client-network-diagnostics`  
> Issue: https://github.com/Zilola5404/StreamPass/issues/4

---

## Что сделано

### Timeouts
- `TimedHttpClient` + `ApiTimeouts.http` (12s) для `StreamPassApi` / `AuthService`
- Splash gate: `isLoggedIn` ограничен `ApiTimeouts.authGate` (5s); при timeout — локальная сессия → Home

### Home CTA (Issue #4 §8)
| Состояние | Текст |
|-----------|--------|
| Disconnected | **Подключить** |
| Connecting | **Подключение…** |
| Connected | **Подключено** |
| Disconnecting | **Отключение…** |
| Server / API fail | **Сервер недоступен** |
| Generic fail | **Не удалось подключиться** |

### Error mapping
- `user_facing_errors.dart` — без технических `SocketException` / `timeout` / `relay_unavailable` в orb

### Connect gate
- По-прежнему: healthy relay + subscription до `VpnChannel.connect`
- При API fail prerequisites → **Сервер недоступен**, TUN не стартует

### Auto Connect
- Макс. **2** неудачи за сессию; дальше только ручной тап «Подключить»

### Windows no-blackhole
- Go `attachRelayAsync`: при `relay_unavailable` → `r.stop()` (TUN + stale routes)
- Flutter: core `error` → `_failSession` / reset session
- Traffic-ready watchdog **35s** → teardown + «Не удалось подключиться»

### Android
- PrepareRelay-before-TUN без изменений (уже соответствовал Issue #4)

---

## Файлы
- `client/lib/services/api_timeouts.dart` (new)
- `client/lib/services/user_facing_errors.dart` (new)
- `client/lib/main.dart`, `streampass_api.dart`, `auth_service.dart`
- `client/lib/screens/home_screen.dart`, `widgets/connect_orb.dart`
- `client/lib/services/windows_traffic_engine.dart`
- `client/go_core/desktop/runtime.go`
- `client/test/user_facing_errors_test.dart`

---

## Сборка
- Android: `app-arm64-v8a-release.apk`
- Windows: `streampasscore.exe` (`-tags with_gvisor`) + Flutter Windows debug/release

## DoD (код)
См. чеклист Issue #4 — пункты про физический E2E остаются на ручной проверке.
