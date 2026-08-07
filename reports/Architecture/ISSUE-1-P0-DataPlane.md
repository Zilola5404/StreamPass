# Architect P0 — GitHub Issue #1 (mirror)

> Источник: https://github.com/Zilola5404/StreamPass/issues/1  
> Создано: 2026-08-07 | Labels: `bug`, `P0`  
> Локальное зеркало для Developer (issue не в git commits)

## Связь с BL-001 (Team Lead)

| | Team Lead BL-001 | Architect Issue #1 |
|--|------------------|-------------------|
| Фокус | Hysteria2 transport re-validation / hardening | Полный **data plane** DIRECT/RELAY + Split по `07.4` |
| Handshake enough? | Нет — нужен real TCP/UDP transfer | Нет — «Connected» ≠ traffic ready |
| Запреты | Нет silent RELAY→DIRECT, нет UDP/443→DIRECT, нет rewrite transport | То же + нет DefaultMode=RELAY, нет широких CIDR hacks, Network Mode не в E05 |
| Device E2E | Обязателен для Done | Обязателен; **Этап 0 = baseline matrix до правок** |

Практический порядок:

1. **Этап 0 Issue #1** — baseline на телефоне (без кода, только evidence).
2. **BL-001** — transport path: protect, TCP/UDP через Hysteria, lifecycle.
3. **Issue #1 этапы 1–10** — DIRECT path, Split, DNS/HostForIP, app bypass, MTU — чинить только подтверждённые дефекты.

Полный текст DoD / QA matrix — в GitHub issue #1.
