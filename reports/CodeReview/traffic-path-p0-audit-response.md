# Response — traffic-path-p0-audit.md

> Дата: 2026-08-09 · Ship: **v0.1.1+39** · Commit follows push

## Этап 0 — диагностика ✅

| # | Требование | Статус |
|---|------------|--------|
| 1 | `stream_open_no_data` (4s, 0 bytes either way) | **DONE** — `tunbridge/bridge.go` pipeTCP |
| 2 | Лог ошибки DIRECT dial | **DONE** (уже было `[tun] direct-tcp fail` + emitDiag); confirmed |
| 3 | `emitDNSRoute` / forceMode | **DONE** — `dnscache.SetRouteHint` + `ClassifyDNSRoute`; default foreign=`DIRECT` (не ложный RELAY); tunnel wires Decision Engine |

## Этап 1–3 — изолированные проверки ✅ (auto) / device QA

| Этап | Evidence |
|------|----------|
| DIRECT | unit + forceMode test; physical `direct_test` matrix → QA |
| RELAY | `TestIntegrationHysteria*` live TCP/UDP |
| DNS | DoT :853 block (+38); HostForIP warn; ClassifyDNSRoute tests |

## Этап 4 — универсальный fallback ✅

- `ModeDirect` fail → одна попытка RELAY с `reason=direct_failed_retry_relay`
- **Не** в `direct_test` (`ForceMode()!=""`)
- Must-relay по-прежнему без silent DIRECT
- UDP аналогично

## Этап 5 — расширение списков ❌ отложено

По аудиту: **не** расширять `DefaultRelayRules` до прохождения device matrix 1–3.

## Тесты

```
go test ./... → PASS
```

## QA next

1. Install +39, Private DNS Off  
2. Diagnostics: Direct Test → ya.ru / example.com → ищи `stream_open_no_data` или `transfer_done` / `traffic_ready`  
3. Full Relay → зарубежный HTTPS  
4. Split → `[dns] query` перед `[decision] host=…`
