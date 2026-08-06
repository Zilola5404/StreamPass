# StreamPass — DIRECT vs VPN bypass (диагностика)

> Дата: 2026-08-06 | Статус: подтверждено кодом + исправлено в клиенте `0.1.1+25`

## Вердикт по разбору

Разбор **в целом верный**. Ниже — согласие / уточнения и что сделано.

| Утверждение разбора | Оценка | Комментарий |
|---|---|---|
| Проблема на Android-клиенте, не в backend rules | ✅ Согласен | `/rules` уже содержит `*.ru` DIRECT |
| Domain DIRECT ≠ снятие `TRANSPORT_VPN` | ✅ Согласен | Зафиксировано в `18_KnownLimitations.md` |
| Банки/Госуслуги проверяют VPN-профиль ОС | ✅ Согласен | Нужен `addDisallowedApplication` |
| Cloudflare DoH для всех доменов ломает RU geo-DNS | ✅ Согласен | Исправлено: split DNS |
| `FLAG_SYSTEM` пропускал предустановленные банки | ✅ Согласен | Исправлено |
| Domain rules не матчят IP-only потоки в TUN | ✅ Согласен | DefaultMode=DIRECT + RU `excludeRoute` |
| Kind=APP на backend как обязательный MVP-фикс | ⚠️ Не согласен как блокер | ТЗ §6: Domain/CIDR/User Rules. Список пакетов надёжнее собирать на устройстве; Kind=APP — опционально позже |
| «Убрать полный VPN» без VpnService | ❌ Нереалистично на Android | Ускоритель foreign-трафика всё равно требует `VpnService`; RU выводится split-tunnel + app-bypass |

## Как это устроено (3 разных механизма)

```
ТЗ «DIRECT» ──► на Android это НЕ одна кнопка, а три слоя:

1) Decision Engine (Go)     — не слать через Hysteria relay
2) Route split (VpnService) — RU IPv4 не захватывать в TUN (excludeRoute / intl-only)
3) App bypass (VpnService)  — пакет приложения исключить из VPN-профиля
                              ← ЕДИНСТВЕННОЕ, что снимает TRANSPORT_VPN
```

Пока приложение **не** в `addDisallowedApplication`, оно видит активный VPN даже если пакеты идут «напрямую».

## Что было сломано

1. **Неверные package id** (до `+22`): Госуслуги = `ru.rostel`, S7 = `ru.s7tl.app`, Мой Налог = `com.gnivts.selfemployed`.
2. **Эвристика пропускала system apps** (`FLAG_SYSTEM`) — предустановленные банки не попадали в bypass.
3. **Весь DNS через Cloudflare DoH** — для `.ru` отдавались не-RU IP → сайт/банкинг мог попасть в TUN и ломаться.
4. **Silent fail** `dialDirectTCP` — ошибки dial не логировались.

## Что исправлено в `0.1.1+25`

- TCP underlay fallback (BL-017): connect candidate logged in diagnostics (`udp/443`, `tcp/8443`, …)
- Secure token storage (audit S-05)

## Что исправлено в `0.1.1+23`

| Фикс | Где |
|---|---|
| Correct package ids + эвристика без слепого skip system | `VpnBypassApps.kt` |
| Split DNS: `.ru/.su/.рф` → Yandex `77.88.8.8`, foreign → DoH | `dnscache/doh.go`, `russian.go` |
| VPN DNS servers → Yandex | `StreamPassVpnService.kt` |
| Лог `[tun] direct-tcp fail` / `relay-tcp fail` | `tunbridge/bridge.go` |
| Документация | этот файл + `18_KnownLimitations.md` |

## Как проверить на устройстве

1. Установить APK `StreamPass-v0.1.1+25-signed-arm64.apk`.
2. Подключить StreamPass → в connect log искать:
   - `VPN app-bypass applied=N` — N ≥ 1 для установленных Госуслуг/ФНС/S7
   - `split-tunnel mode=exclude-ru ruExcludes=...` (Android 13+)
   - `[dns] query gosuslugi.ru via=yandex`
3. Открыть Госуслуги / Мои налоги / S7 — не должно требовать «отключите VPN».
4. В браузере `2ip.ru` / `yandex.ru` — российский IP (не IP relay).
5. YouTube / Instagram — через relay (ускорение).

## Ограничения, которые остаются

- Иконка ключа VPN в статус-баре **останется** при активном StreamPass — это Android `VpnService`, не баг маршрутизации.
- Chrome / системный браузер **нельзя** целиком выкинуть из VPN (иначе YouTube в браузере не ускорится). RU-сайты в браузере опираются на split DNS + RU CIDR exclude.
- Kind=APP на backend — backlog (remote update списка пакетов без релиза APK).

## Backend

Менять backend **не требуется** для этой проблемы. Rules v4 с `*.ru` DIRECT достаточны. Проблема была в OS-level видимости VPN и DNS на клиенте.
