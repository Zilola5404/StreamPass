# Current Task

> Updated: 2026-08-04

## Главная задача

**BL-014: Exclusions sync** (следующий после BL-006)

## Описание

Синхронизация пользовательских исключений с backend при изменении в настройках.

## Previous Task (Completed)

**BL-006: Rule Engine (клиент)** — 2026-08-04

- `RuleEngineService` — polling по `rule_poll_interval_sec`, hot-reload через `VpnChannel.updateRules`
- Go: `UpdateRules`, `ActiveRulesVersion`, `AtomicEngine`
- Android: MainActivity → StreamPassVpnService → TunnelBridge → mobile.UpdateRules
- Build: **v0.1.1+6 (rule-engine-bl006-v1)**

## Deferred from BL-006

- Backend sync для user exclusions → **BL-014**
