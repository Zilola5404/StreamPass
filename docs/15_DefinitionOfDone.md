# StreamPass — Definition of Done

> Дата: 2026-08-03

---

## Критерии готовности задачи

Задача считается **Done**, когда выполнены ВСЕ применимые пункты:

### 1. Код написан

- [ ] Реализация соответствует ТЗ и Clean Architecture
- [ ] `go build ./...` — успешно (backend/shared)
- [ ] `go vet ./...` — без замечаний
- [ ] `flutter analyze` — без ошибок (если client)
- [ ] Нет hardcoded secrets
- [ ] Изменения минимальны по scope (не over-engineer)

### 2. Тесты проходят

- [ ] `go test ./...` — green
- [ ] `flutter test` — green (если client)
- [ ] Новая бизнес-логика покрыта unit-тестами
- [ ] Существующие тесты не сломаны

### 3. Документация обновлена

- [ ] `docs/10_Progress.md` — запись о задаче
- [ ] `docs/04_Backlog.md` — статус задачи → Done
- [ ] `docs/08_API.md` — если изменён API
- [ ] `docs/09_Database.md` — если изменена схема
- [ ] `docs/07_Architecture.md` — если изменена архитектура
- [ ] `docs/11_Decisions.md` — если архитектурное решение
- [ ] `docs/17_CHANGELOG.md` — если user-facing change
- [ ] `ai/LastSession.md` — итог AI-сессии

### 4. Security проверен

- [ ] Нет новых security vulnerabilities
- [ ] Auth endpoints защищены (rate limit, validation)
- [ ] Secrets только через env
- [ ] Telemetry не содержит PII
- [ ] Input validation на всех endpoints

### 5. Performance проверен

- [ ] Нет очевидных N+1 queries
- [ ] Connection pool settings адекватны
- [ ] Rate limiting не сломан
- [ ] Client: нет blocking UI operations

### 6. Code review выполнен

- [ ] Self-review diff перед commit
- [ ] Соответствие `docs/21_CodingStandards.md`
- [ ] Соответствие `docs/22_NamingConvention.md`
- [ ] Нет unrelated changes в diff

### 7. Smoke test

- [ ] `GET /health` → 200
- [ ] Auth flow работает (register → login → protected endpoint)
- [ ] Docker compose up (если infra changes)
- [ ] Android app запускается (если client changes)

---

## Уровни Done

| Уровень | Когда |
|---------|-------|
| **Task Done** | Все пункты выше (применимые) |
| **Sprint Done** | Все tasks sprint goal выполнены |
| **Release Done** | См. `docs/12_ReleasePlan.md` checklist |
| **MVP Done** | См. `docs/02_TZ.md` §22 критерии готовности |

---

## Исключения

- Документация-only задачи: пункты 1, 2, 5, 6 не применимы
- Infra-only задачи: пункт 5 — load test optional
- Spike/POC: согласовать reduced DoD заранее
