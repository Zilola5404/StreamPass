# Current Task

> Updated: 2026-08-03

## Главная задача

**BL-005: Decision Engine (клиент)**

## Описание

Реализовать на клиенте Decision Engine: маршрутизация DIRECT / RELAY / FALLBACK per-connection по правилам ТЗ §5.

## Контекст

- BL-003 завершён: transport verified (integration tests + APK), TunnelBridge fix (`mobile.Mobile`)
- Production relay 212.43.159.198 был недоступен; локальный test relay в `docker-compose.hysteria-test.yml`
- Physical Android TUN E2E — manual when device available (see `reports/BL-003-test-report.md`)

## Acceptance Criteria

- [ ] Decision Engine выбирает DIRECT vs RELAY per connection
- [ ] Интеграция с Rule Engine (BL-006) или минимальные правила
- [ ] Не ломает текущий full-tunnel MVP path

## Previous Task (Completed)

BL-003 End-to-end VPN verification — completed 2026-08-03.
