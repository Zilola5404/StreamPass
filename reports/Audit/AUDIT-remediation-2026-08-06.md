# Audit remediation — 2026-08-06

> Закрытие открытых пунктов из `reports/Audit/*` и `reports/CodeReview/2c56972-1-review.md`

---

## Статус по источникам

| Источник | Вердикт | Действие |
|----------|---------|----------|
| AUDIT-001 | INVALID (wrong target) | Без изменений кода — см. `AUDIT-001-response.md` |
| AUDIT-003 | Закрыт ранее (+15) | BUG-001…004 исправлены; ответ в `AUDIT-003-response.md` |
| BL-005 audit (2026-08-03) | Устарел | BL-005/006/016/017 Done; см. `BL-005-audit-response.md` |
| Code review `2c56972` | P0/P1 закрыты в `9390d10` | P2/P3 — этот коммит |

---

## Исправлено в этом проходе

| ID | Замечание | Fix |
|----|-----------|-----|
| CR-P3 | Коллизия ADR-010/011 | Переименованы в **ADR-012/013**, обновлены ссылки |
| CR-P2-3 | Дубли auth-expired эвристики | `AuthErrorCodes` — единый модуль для API + VPN UI |
| CR-P2-4 / S-05 | Токены в SharedPreferences | `TokenStorage` + `flutter_secure_storage`, миграция legacy keys |
| CR-P2-5 | ConnectionLog O(n) trim | `Queue.removeFirst()` |
| S-04 | Webhook без доп. проверки | Optional `billing.webhook_secret` + header `X-StreamPass-Webhook-Secret` |
| BL-005 / P0 secrets | Пароли в scripts/docs | `register-region-relays.sh` требует env; `docs/02_TZ.md` — placeholders |
| — | Тесты после `TokenStorage` | `TokenStorage.inMemory()` в widget/E2E; миграция prefs не удаляет ключ при failed secure write |
| — | `VpnChannel.connect` в unit-тестах | Валидация `connection_config` до `ensureListening()` |

---

## Проверка (2026-08-06)

- `client`: **49/49** `flutter test` — PASS
- `backend`: `go test ./...` — PASS

| ID | Пункт | Почему не в этом diff |
|----|-------|------------------------|
| S-02 | `connection_config` plaintext в PostgreSQL | Нужна migration + key management |
| BL-023…025 | Windows/iOS/macOS | Вне scope MVP |
| BL-004 | ЮKassa live | Skipped по решению продукта |
| S-09 | VPS hardening checklist | Ops/manual на серверах |
| S-08 | Certificate pinning | Post-MVP client hardening |

---

## Остаётся (осознанный backlog)
