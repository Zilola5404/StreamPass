# Current Stage

> Дата: 2026-08-03

## Текущий этап разработки

**Phase 1: MVP Completion** — Backend done, Client UI done, VPN tunnel implemented (E2E test pending).

## Цель этапа

Пользователь может: зарегистрироваться → оплатить → нажать «Подключить» → получить рабочий VPN через Hysteria2 relay с автоматической маршрутизацией.

## Что должно быть готово

| Component | Required | Status |
|-----------|----------|--------|
| Backend API (all modules) | ✅ | Done |
| PostgreSQL + Redis + Docker | ✅ | Done |
| Flutter Android UI (all screens) | ✅ | Done |
| Auth integration (client ↔ backend) | ✅ | Done |
| Hysteria2 tunnel (go_core) | ✅ | Done (BL-001) |
| streampasscore.aar + Android integration | ✅ | Done (BL-002) |
| E2E VPN on device | ✅ | Open (BL-003) |
| Decision Engine (client) | ✅ | ❌ Not started |
| Rule Engine (client) | ✅ | ❌ Not started |
| ЮKassa live test | ✅ | ❌ Not tested |
| CI/CD | Nice to have | ❌ Not started |

## Критерии завершения этапа

1. VPN connect works end-to-end on real Android device
2. Foreign IP verified (ifconfig.me shows relay IP)
3. Russian domains route direct (when Rule Engine implemented)
4. Payment flow tested with ЮKassa sandbox
5. All P0 backlog items (BL-001 through BL-006) closed
6. See `docs/30_FinalAcceptance.md` for full criteria

## Estimated Completion

TODO: Требуется уточнение после BL-003 (device test)
