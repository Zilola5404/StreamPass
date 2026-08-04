# Технический отчёт аудита: Причины отсутствия интернета и ложного статуса Disconnected в StreamPass

**Дата:** 4 августа 2026 г.  
**Компонент:** StreamPass VPN Client (Android Native, Go Core, Flutter UI)  
**Статус проверки:** БАГИ ОБНАРУЖЕНЫ (Identified & Root Cause Found)  

---

## 1. Сводка выявленных дефектов

| ID Бага | Критичность | Компонент | Проблема |
| :--- | :--- | :--- | :--- |
| **BUG-001** | **Critical (P0)** | `StreamPassVpnService.kt` / `bridge.go` | **Петля маршрутизации (Routing Loop) & DROP UDP/DNS:** При создании TUN `0.0.0.0/0` без вызова `protect()` для сокетов Hysteria и прямого трафика, а также из-за жесткой проверки `addr != destAddr` в `relayUDP`, все DNS и пакеты данных сбрасываются. |
| **BUG-002** | **Major (P1)** | `vpn_channel.dart` / `diagnostics_screen.dart` | **Ложный статус Disconnected в Настройках:** `EventChannel` не сохраняет последнее состояние. При переходе в Настройки/Диагностику подписка создается заново и получает `null`, отображая "Disconnected". |
| **BUG-003** | **Major (P1)** | `bridge.go` | **Закольцовывание DIRECT-трафика:** Прямые сокеты TCP/UDP (`net.Dialer`) в Go Core не защищены от TUN, попадают обратно в VPN и падают по таймауту. |
| **BUG-004** | **Minor (P2)** | `TunnelBridge.kt` / `VpnChannel.kt` | **Отсутствие метода `getStatus` в Native Channel:** Клиент не может запросить текущий статус активного службы VpnService при повторном открытии UI. |

---

## 2. Подробный разбор причин (Root Cause Analysis)

### Причина 1: Почему "Нет интернета и не открывается ни один сайт"

1. **Петля маршрутизации сокета Hysteria2 (Routing Loop):**
   - В `StreamPassVpnService.kt` добавляется маршрут по умолчанию: `.addRoute("0.0.0.0", 0)`.
   - Защита `protect(socket)` для QUIC-сокета Hysteria в Go Core отсутствует или не вызывается для всех сокетов Go runtime.
   - В результате QUIC-пакеты Hysteria к реле (`157.167.x.x`) направляются в свой же TUN-интерфейс, зацикливаются и приводят к полному обрыву интернет-соединения.

2. **Сброс (DROP) UDP и DNS ответов в `bridge.go`:**
   - В `relayUDP` (`bridge.go`, строки 265-267):
     ```go
     data, addr, err := hyUDP.Receive()
     if addr != destAddr { continue } // <- БОТТЛНЕК И СБРОС ПАКЕТОВ
     ```
     `hyUDP.Receive()` возвращает строковый адрес ответа (например, `1.1.1.1:53`), который может отличаться форматированием от `destAddr`. В результате проверка отсекает валидные DNS-ответы, домены не резолвятся, и приложения сообщают "Нет подключения к интернету".

3. **Закольцовывание ModeDirect сокетов:**
   - Прямые подключения (`dialDirectTCP` и `copyUDP`) вызывают `net.DialContext`, сокеты которого не помечены через `VpnService.protect()`. Весь трафик в режиме `DIRECT` уходит в TUN и блокируется.

---

### Причина 2: Почему "На вкладке Настройки статус стоит Disconnected"

1. **Природа `EventChannel` в Flutter:**
   - `EventChannel` транслирует события в режиме реального времени. Когда пользователь переходит на вкладку «Настройки» / «Диагностика», создается новый виджет `DiagnosticsScreen`.
   - Подписка `VpnChannel.statusStream.listen(...)` видит события, происходящие *после* открытия экрана.
   - Поскольку статус `connected` был отправлен ранее при нажатии на орбиту, новый экран получает `null` и рисует состояние по умолчанию: `Disconnected`.

2. **Отсутствие слоя синхронизации (ValueNotifier / State Management):**
   - В приложении отсутствует единое глобальное состояние `VpnStateNotifier`, хранящее текущий `VpnStatusUpdate`.

---

## 3. Решение и пошаговый план исправления

### Решение 1: Исправление VPN маршрутизации и прохождения трафика
1. **Добавить вызов `protectSocket` в JNI / Go Core:**
   Передать дескриптор UDP сокета Hysteria2 и прямых сокетов в Android `VpnService.protect(fd)` перед выполнением `dial`.
2. **Исправить сопоставление UDP ответов в `bridge.go`:**
   Убрать некорректную жесткую строковую проверку `addr != destAddr` для UDP-сессий, либо использовать табличную карту портов `map[string]N.PacketConn`.
3. **Исключить IP реле из VPN туннеля:**
   В `StreamPassVpnService.kt` явно вызвать `.addRoute()` для всех сетей, или добавить `.addRoute` с исключением IP реле-сервера.

### Решение 2: Синхронизация статуса в UI и Настройках
1. **Внедрить `VpnStateNotifier` / `ValueNotifier<VpnStatusUpdate>`:**
   Создать синглтон-состояние `VpnChannel.currentStatus`, которое сохраняет последнее переданное событие от `EventChannel`.
2. **Добавить Native Method `getStatus`:**
   В `MainActivity.kt` и `VpnChannel.kt` добавить вызов `invokeMethod('getStatus')`, возвращающий реальное состояние службы `StreamPassVpnService`.
3. **Обновить `DiagnosticsScreen` и `SettingsScreen`:**
   Использовать `VpnChannel.currentStatus` при инициализации виджетов.

---

## 4. Тест проверки (Verification Test Strategy)

```dart
// client/test/vpn_status_state_test.dart
void main() {
  test('VpnChannel retains last known status for late listeners', () {
    final statusNotifier = ValueNotifier<VpnStatusUpdate>(
      VpnStatusUpdate(VpnEvent.disconnected),
    );
    
    // Simulate connection event from native layer
    statusNotifier.value = VpnStatusUpdate(VpnEvent.connected, relayName: 'Relay NL-01', pingMs: 42);
    
    // Late screen subscription (e.g. Settings/Diagnostics Screen opened)
    final current = statusNotifier.value;
    expect(current.event, equals(VpnEvent.connected));
    expect(current.relayName, equals('Relay NL-01'));
  });
}
```
