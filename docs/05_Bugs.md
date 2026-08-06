# StreamPass — Журнал ошибок

> Дата: 2026-08-05

---

## BUG-001: VPN tunnel — stub, подключение не работает

| Поле | Значение |
|------|----------|
| **Описание** | Go core tunnel возвращает ошибку stub вместо реального Hysteria2 подключения |
| **Как воспроизвести** | 1. Запустить Android app 2. Нажать Connect 3. VPNService вызывает TunnelBridge → go_core |
| **Причина** | `client/go_core/mobile/tunnel.go` — intentional stub, transport не реализован |
| **Решение** | Реализован Hysteria2 client, AAR собран (BL-001, BL-002, BL-003) |
| **Статус** | Fixed / Closed — 2026-08-03 |
| **Файлы** | `client/go_core/mobile/tunnel.go`, `client/android/.../TunnelBridge.kt` |

---

## BUG-002: README — неверный статус Health Monitor

| Поле | Значение |
|------|----------|
| **Описание** | README.md утверждает, что Health Monitor worker не реализован |
| **Как воспроизвести** | Прочитать README.md § «Что НЕ реализовано» |
| **Причина** | README не обновлён после добавления `backend/cmd/healthmonitor/` |
| **Решение** | Обновлён README (BL-012) |
| **Статус** | Fixed / Closed — 2026-08-04 |
| **Файлы** | `README.md` |

---

## BUG-003: 03_CurrentState.md содержал ложные данные

| Поле | Значение |
|------|----------|
| **Описание** | Документ утверждал «OAuth готово», «Telegram частично», «Payments не реализовано» — не соответствует коду |
| **Как воспроизвести** | Сравнить старый 03_CurrentState.md с кодовой базой |
| **Причина** | Шаблон заполнен без анализа кода |
| **Решение** | Исправлено в текущей версии 03_CurrentState.md |
| **Статус** | Fixed / Closed — 2026-08-03 |
| **Файлы** | `docs/03_CurrentState.md` |

---

## BUG-004: streampasscore.aar отсутствует

| Поле | Значение |
|------|----------|
| **Описание** | TunnelBridge.kt ловит ClassNotFoundException если AAR не в libs |
| **Как воспроизвести** | Сборка Android без предварительной сборки go_core AAR |
| **Причина** | AAR не собран и не добавлен в репозиторий (by design) |
| **Решение** | AAR собран и подключён (BL-002); см. `client/go_core/README.md` |
| **Статус** | Fixed / Closed — 2026-08-03 |
| **Файлы** | `client/android/.../TunnelBridge.kt`, `client/android/app/libs/` |

---

## BUG-005: Release signing — debug keys

| Поле | Значение |
|------|----------|
| **Описание** | Android release build использует debug signing keys |
| **Как воспроизвести** | `flutter build apk --release` без `key.properties` |
| **Причина** | TODO в build.gradle.kts не выполнен |
| **Решение** | Production keystore path via `key.properties` (BL-013); APK v0.1.1+25 signed when JKS present |
| **Статус** | Fixed / Closed — 2026-08-04 |
| **Файлы** | `client/android/app/build.gradle.kts` |

---

## BUG-006: vendor-src/mobile не найден

| Поле | Значение |
|------|----------|
| **Описание** | go_core/go.mod ссылается на `../vendor-src/mobile` — директория отсутствует |
| **Как воспроизвести** | `cd client/go_core && go build ./...` |
| **Причина** | gomobile dependency не vendored |
| **Решение** | `vendor-src/mobile` добавлен; go_core собирается |
| **Статус** | Fixed / Closed — 2026-08-04 |
| **Файлы** | `client/go_core/go.mod`, `vendor-src/mobile/` |

---

## Шаблон для новых багов

```
## BUG-XXX: [Краткое описание]

| Поле | Значение |
|------|----------|
| **Описание** | |
| **Как воспроизвести** | |
| **Причина** | |
| **Решение** | |
| **Статус** | Open / In Progress / Fixed / Won't Fix |
| **Файлы** | |
```
