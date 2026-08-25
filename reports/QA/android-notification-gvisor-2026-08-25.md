# Android: кликабельное VPN-уведомление + AAR с gVisor

> Дата: 2026-08-25  
> Ветка: `fix/client-network-diagnostics`  
> Устройство (проверка): физический Android, Connect после переустановки APK

---

## 1. Кликабельное уведомление VPN

### Проблема
Foreground-уведомление StreamPass («Подключено» / «Подключение…») отображалось, но **не открывало приложение** по тапу — у `NotificationCompat.Builder` не было `setContentIntent`.

### Решение
В `StreamPassVpnService.kt`:
- `openAppPendingIntent()` → `MainActivity` с `ACTION_MAIN` + `CATEGORY_LAUNCHER`, флаги `NEW_TASK | SINGLE_TOP | CLEAR_TOP`
- `PendingIntent.FLAG_UPDATE_CURRENT | FLAG_IMMUTABLE` (API 23+)
- `setContentIntent` на статусное и failure-уведомления
- `setOnlyAlertOnce(true)` на статусе, чтобы не дёргать при обновлении текста

### Файлы
- `client/android/app/src/main/kotlin/com/streampass/app/StreamPassVpnService.kt`

---

## 2. Connect fail: «gVisor is not included in this build»

### Симптом
После Connect уведомление с текстом:

```text
tun bridge: gVisor is not included in this build, rebuild with -tags with_gvisor
```

### Причина
`tunbridge` всегда запрашивает stack `gvisor` (см. комментарий в `bridge.go` — `system` stack ломает TCP на One UI 8 / Android 16).  
Последний `streampasscore.aar` (ReconnectRelay / Issue #2) был собран **без** `-tags with_gvisor`, поэтому sing-tun отдавал stub-ошибку.

### Решение
```bash
cd client/go_core
gomobile bind -tags with_gvisor -target=android -androidapi=21 -o streampasscore.aar ./mobile
cp streampasscore.aar ../android/app/libs/
```

Размер AAR: ~37.5 MB (с gVisor). В `libgojni.so` (arm64) есть ссылки на gvisor.

### Документация сборки
Обновлены команды (обязательный тег `with_gvisor`):
- `README.md`
- `client/README.md`
- `client/go_core/README.md`

---

## 3. APK

```text
client/build/app/outputs/flutter-apk/app-arm64-v8a-release.apk
```

Сборка: `flutter build apk --release --split-per-abi --target-platform android-arm64`  
(локальный артефакт, в Git не коммитится)

### Проверка на устройстве
1. Установить arm64 release APK  
2. Connect → уведомление «Подключено»  
3. Тап по уведомлению → открывается StreamPass  
4. Нет ошибки про `with_gvisor`

---

## Связь с Issue #2

AAR с `ReconnectRelay` остаётся актуальным; пересборка только добавила обязательный build tag `with_gvisor`, без которого Android TUN не стартует.
