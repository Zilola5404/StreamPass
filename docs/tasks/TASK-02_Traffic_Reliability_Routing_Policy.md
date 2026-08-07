# TASK-02: Traffic Reliability — REWORK по Routing Policy

> Статус: **Ready for Developer (REWORK)** · 2026-08-07  
> Политика: `docs/07.4_RoutingPolicy.md`  
> ADR: `docs/11_Decisions.md` → ADR-015  
> Architecture: `reports/Architecture/TASK-02-ArchitectureDecision.md`

## Цель

Привести клиентскую маршрутизацию к ТЗ: RU → DIRECT, foreign → RELAY; убрать product-решения из Transport; починить data-plane RELAY без маскировки DIRECT.

## Не делать

- `quic_direct_bypass` в default `split`
- Silent «всегда RELAY fail → DIRECT»
- `DefaultMode=RELAY`
- Network Mode на E05 как user feature
- Cloudflare `/12` в builtin defaults

## Сделать

См. чеклист REWORK в Architecture Decision ADDENDUM + §11 в `docs/07.4_RoutingPolicy.md`.

## DoD

§12 в `docs/07.4_RoutingPolicy.md` + device QA matrix.
