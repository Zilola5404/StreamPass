# BL-005 Audit — Response

> **Audit:** `reports/Audit/BL-005-audit.md` (commit `2c56972`, score 48/100, FAIL)  
> **Response date:** 2026-08-04  
> **Current commit:** `bcfb491+` (BL-005/BL-006 implemented)

---

## Summary

Аудит выполнен **до** merge BL-005/BL-006. Часть критических замечаний **устранена** в `bcfb491`. Остальное — backlog MVP или осознанные компромиссы.

---

## Принято и исправлено

| Audit item | Action |
|------------|--------|
| Decision Engine отсутствует (P0) | **Done** — `client/go_core/internal/decision/`, интеграция в `tunbridge`, BL-005 |
| Rule Engine на клиенте (P0) | **Done** — `RuleEngineService`, `UpdateRules`, BL-006 |
| TUN packet dispatcher (P0 arch) | **Частично закрыто** — `sing-tun` system stack + `routingHandler` per-flow (не отдельный gVisor fork, но парсинг IP/TCP/UDP из TUN fd реализован) |
| Секреты в docs/scripts (security) | **Fixed** — placeholders в `docs/RelayServers.md`, тестах, `setup-relay-hysteria.sh` |
| CI/CD отсутствует (P1) | **Fixed** — `.github/workflows/ci.yml` |
| Word/docx в репозитории | **Fixed** — удалены из Git (см. commit) |
| `android_old/` мёртвый код | **Fixed** — удалён из Git |

---

## Не согласны / отложено (без изменения кода)

| Audit item | Почему |
|------------|--------|
| **FAIL вердикт «нет Decision Engine»** | Устарел: BL-005/006 merged. Аудит смотрел `2c56972`. |
| **90% Go-ядро на клиенте** | MVP: UI/auth/settings остаются на Flutter; routing/tunnel — Go. Полный перенос Config/Telemetry/AutoUpdate в go_core — roadmap, не блокер BL-005. |
| **Windows/macOS/iOS** | ТЗ §21 MVP exclusions + `docs/18_KnownLimitations.md` — только Android в v0.1.x. |
| **DNS Cache / DoH (§7)** | Backlog **BL-016**; hardcoded 1.1.1.1 в VpnService — временно. |
| **Fallback ports (§10)** | Backlog **BL-017**; один `connection_config` от backend — осознанный MVP scope. |
| **Шифрование `connection_config` в PostgreSQL** | Валидный риск; требует backend migration + key management. Backlog security; не scope code-review fix. |
| **Удалить `streampasscore.aar` из Git** | **Не согласны с удалением сейчас:** gomobile CI ещё не собирает AAR на каждый PR; без AAR Android-сборка ломается offline. `.gitignore` исключает дубликат в `go_core/`, tracked copy — `android/app/libs/` для reproducible builds. Цель — CI-built AAR позже, не немедленное удаление. |
| **Ручной JWT/YAML/Redis RESP** | ADR-002: dependency-free infra для sandbox. Замена на библиотеки — отдельная задача при выходе из sandbox. |
| **Refresh token rotation** | Backlog **BL-015**; базовый `/refresh` уже есть. |
| **Backlog BL-001/003 «Done» без device E2E** | Согласны с риском; статус Done = transport code + integration tests PASS. E2E на устройстве — manual QA, отражено в `reports/BL-003-test-report.md`. |

---

## Связанный code review

Замечания `reports/CodeReview/2c56972-review.md` обработаны отдельно (pingMs, secrets, docs, refresh dedup, ADR-010/011).

---

## Рекомендуемый пересмотр аудита

После установки APK **v0.1.1+6** на устройство проверить:

1. Diagnostics → build label `rule-engine-bl006-v1`
2. Connect → ping > 0 ms (после fix pingMs)
3. Rules hot-reload при изменении version на backend

Ожидаемый пересмотренный статус BL-005 scope: **PASS with known MVP limitations** (не full TZ compliance).
