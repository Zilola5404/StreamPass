# Ответ на AUDIT-003 — VPN diagnostics / no internet / Disconnected

**Дата:** 2026-08-04  
**Исходный аудит:** `reports/Audit/AUDIT-003-vpn-diagnostics.md`  
**Код на момент ответа:** клиент `v0.1.1+15`

---

## Сводка

| ID | Вердикт | Действие |
|----|---------|----------|
| BUG-001 protect / routing loop | **Частично устарело** | `protect()` уже был; корень «нет интернета» в +14 — TUN `10.10.0.2/30` (broadcast NAT). Protect оставляем. |
| BUG-001 UDP `addr != destAddr` | **Согласен** | Исправлено: `sameUDPEndpoint()` (+ тесты). |
| BUG-001 «DNS дропается этой проверкой» | **Не согласен** | UDP/53 идёт в `handleDNS`/DoH, не в `relayUDP`. |
| BUG-002 lastStatus | **Согласен** | Уже было в +14 (`VpnChannel.lastStatus`); усилено. |
| BUG-003 DIRECT без protect | **Устарело** | `protect.Control` уже на DIRECT TCP/UDP и DoH. |
| BUG-004 `getStatus` | **Согласен** | Добавлен MethodChannel `getStatus` + `fetchNativeStatus()`. |
| «Исключить relay через addRoute» | **Не согласен** | На Android правильный путь — `VpnService.protect(fd)`, не exclude через `addRoute`. |

---

## Что исправлено сейчас (+15)

1. **UDP reply matching** — строковое `addr != destAddr` заменено на нормализацию IP:port.
2. **`getStatus`** — native snapshot + Diagnostics вызывает `fetchNativeStatus()` при открытии.
3. **Тесты** — `udp_endpoint_test.go`, `vpn_status_state_test.dart`.

---

## С чем не согласен (подробно)

### 1. Protect как единственная причина «нет интернета»
Аудит верно описывает класс багов (routing loop без protect), но на момент симптомов пользователя protect уже был в логах (`protect(fd=…)=true`). Реальный блокер data-plane после handshake:

- DoH recursion на hostname (`+13`)
- TUN prefix `10.10.0.2/30` → `Next()=10.10.0.3` broadcast → сломанный TCP NAT system stack (`+14`)

### 2. DNS якобы режется `addr != destAddr`
В текущем `bridge.go` порт 53 перехватывается до `relayUDP`. Фильтр бил не DNS DoH-путь, а обычный UDP через Hysteria. Фикс всё равно нужен — для QUIC/UDP приложений.

### 3. Исключение IP relay через `addRoute`
`Builder.addRoute` только **добавляет** префиксы в VPN. Исключение underlay — это `protect(fd)` (и опционально API 33+ `excludeRoute`, не как основной механизм). Предложение аудита здесь технически неверно для нашей схемы.

---

## Уже было до этого аудита (не «дыры»)

- SocketProtector + `hyconfig` ConnFactory  
- DIRECT `protect.Control`  
- `VpnChannel.lastStatus` + early EventChannel listen  
- TUN `10.10.0.1/30`
