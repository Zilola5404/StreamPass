# TASK-01: Diagnostic Telemetry & Performance Monitoring

> Файл задачи: `docs/tasks/TASK-01_Diagnostic_Telemetry_Performance_Monitoring.md`  
> Статус на 2026-08-07: **частично выполнено** (MVP операторской диагностики). Полный wishlist из ТЗ ниже размечен.

## Цель задачи

Добавить в StreamPass систему диагностики, которая позволит понять:

1. Через какой маршрут идет трафик: DIRECT / RELAY
2. На какой сайт идет соединение (hostname / `https://host`, без path)
3. Почему выбран RELAY или DIRECT (rule + decision_reason)
4. Где проблема: DNS / dial / relay / slow / cut
5. Почему сайт долго грузится / не открывается

---

## Актуальность vs текущая реализация

| Требование TASK-01 | Статус | Где |
|---|---|---|
| Traffic events (site, IP, proto, route) | ✅ | `diag_events` + `[diag]` / `[route]` в `connect.log` |
| Decision reason (почему DIRECT/RELAY) | ✅ | `DecideDetailed` → `rule` / `decision_reason` |
| Dial latency + slow (>1.5s) | ✅ | `latency_ms`, `slow`, `result=slow` |
| Transfer speed | ✅ | `speed_kbps` после ≥32KB transfer |
| DNS fail/ok | ✅ | `proto=dns` events |
| Connection errors (timeout/refused/cut) | ✅ | `result` + `reason` + `error_code` |
| Upload to backend | ✅ | `POST /api/v1/diag` (не отдельный `/telemetry/events`) |
| Admin view | ✅ | Admin tab **Diagnostics** + `PullDiagLogs.ps1` |
| Debug toggle in Settings | ✅ | «Диагностика трафика» |
| Privacy (no full URL/cookies) | ✅ | только host / `https://host` |
| App package / YouTube UID | ❌ отложено | нет UID из TUN без Android API glue |
| ASN / country / provider | ❌ отложено | нужен GeoIP DB, YAGNI для MVP |
| TLS handshake / TTFB | ❌ отложено | без MITM на TUN не измерить TLS |
| Ping/loss/jitter per flow | ❌ отложено | есть session ping relay, не per-site |
| Отдельные таблицы traffic/decision/perf/error | ❌ не делаем | одна `diag_events` (KISS) |
| Mbps speedtest | ❌ не делаем | есть throughput `speed_kbps` на flow |

**Вывод:** задача актуальна как продуктовый wishlist; для операторской диагностики MVP **достаточен** текущий канал `diag`. Не плодим 4 таблицы и дубль API.

---

## Критерии готовности (DoD)

| Критерий | Статус |
|---|---|
| видно какой сайт | ✅ `site` / `host` (+ reverse DNS IP→host) |
| видно какой IP | ✅ `dest_ip` |
| видно DIRECT или RELAY | ✅ `mode` |
| видно почему выбран маршрут | ✅ `rule` + `decision_reason` |
| видно скорость | ✅ `speed_kbps` (xfer) |
| видно задержку | ✅ `latency_ms` (dial) |
| видно причину медленной загрузки | ✅ `slow` / `slow_dial_*` / `transfer_done` |
| видно причину ошибки | ✅ `timeout_*` / `relay_dial_fail` / … |
| данные в backend dashboard | ✅ Admin → Diagnostics |

---

## Как пользоваться

1. APK с diag uploader, VPN on, открыть сайты.
2. Admin: `https://212-43-156-33.nip.io/admin/` → вкладка **Diagnostics**.
3. Или: `.\scripts\PullDiagLogs.ps1 -OutFile .\diag.json`

Локально на устройстве: `filesDir/connect.log` (`[diag]`, `[route]`, `[decision]`).
