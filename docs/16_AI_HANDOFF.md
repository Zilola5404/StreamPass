# StreamPass — AI Handoff Guide

> Инструкция передачи проекта между AI-сессиями (Cursor, Claude Code, Copilot, Codex)  
> Дата: 2026-08-05

---

## Как новый AI должен начать работу

### Шаг 1: Прочитать обязательные документы (5 мин)

```
1. docs/00_ProjectRules.md     — конституция проекта
2. docs/14_AIContext.md        — полный AI-контекст
3. docs/03_CurrentState.md     — что работает / что нет
4. ai/CurrentTask.md           — текущая задача
5. ai/CurrentFocus.md          — фокус проекта
6. ai/LastSession.md           — что сделал предыдущий AI
```

### Шаг 2: Проверить ai/ folder

| Файл | Содержание |
|------|------------|
| `ai/CurrentStage.md` | Этап разработки |
| `ai/CurrentTask.md` | Главная задача |
| `ai/CurrentSprint.md` | Sprint goal + tasks |
| `ai/CurrentFocus.md` | Фокус |
| `ai/NextTask.md` | Следующая задача |
| `ai/LastSession.md` | Итог последней сессии |
| `ai/OpenQuestions.md` | Нерешённые вопросы |
| `ai/AINotes.md` | Рабочие заметки |

### Шаг 3: Проверить последние изменения

```bash
git log --oneline -10
git diff HEAD~3..HEAD --stat
git status
```

### Шаг 4: Проверить сборку

```bash
go build ./...
go test ./...
cd client && flutter analyze && flutter test
```

### Шаг 5: Начать работу

- Взять задачу из `ai/CurrentTask.md` или `docs/04_Backlog.md`
- Следовать `docs/00_ProjectRules.md`
- По завершении обновить `ai/LastSession.md` и `docs/10_Progress.md`

---

## Что передавать при handoff

### Обязательно

1. **CurrentTask** — что делали, что осталось
2. **Changed files** — список изменённых файлов
3. **Blockers** — что блокирует прогресс
4. **Open questions** — нерешённые вопросы для product/architect
5. **Test status** — green/red

### Шаблон LastSession

```markdown
## Дата: YYYY-MM-DD

### Что сделал AI:
- ...

### Какие файлы изменил:
- path/to/file.go
- ...

### Что осталось:
- ...

### Следующие действия:
- ...

### Blockers:
- ...

### Test status:
- go test: green/red
- flutter test: green/red
```

---

## Карта документов

```
docs/
├── 00_ProjectRules.md    ← START HERE (rules)
├── 01_Project.md         ← project passport
├── 02_TZ.md              ← full product spec
├── 03_CurrentState.md    ← what's implemented
├── 04_Backlog.md         ← task list
├── 05_Bugs.md            ← bug tracker
├── 06_TestPlan.md        ← testing strategy
├── 07_Architecture.md    ← system design
├── 08_API.md             ← API reference
├── 09_Database.md        ← DB schema
├── 10_Progress.md        ← dev history
├── 11_Decisions.md       ← ADR records
├── 12_ReleasePlan.md     ← release phases
├── 13_Risks.md           ← risk register
├── 14_AIContext.md       ← AI context (READ SECOND)
├── 15_DefinitionOfDone.md
├── 16_AI_HANDOFF.md      ← this file
└── 17-32                 ← extended docs

ai/
├── CurrentTask.md        ← what to work on NOW
├── CurrentFocus.md       ← project focus
├── LastSession.md        ← previous session summary
└── ...

prompts/
├── SeniorArchitect.md    ← role prompts
├── BackendDeveloper.md
└── ...
```

---

## Частые ошибки AI (избегать)

1. **Придумывать компоненты** — OAuth, Telegram не существуют; Admin UI **есть** (`admin/`, `/admin/`)
2. **Не обновлять docs** — после каждой задачи обновлять Progress + LastSession
3. **Ломать Clean Architecture** — business logic в handlers
4. **Коммитить без запроса** — только по explicit request
5. **Считать VPN stub** — туннель **работает** (Hysteria2 + AAR); не помечать BL-001…003 как Open
6. **Стартовать BL-023/024/025 или ЮKassa live** без явного запроса пользователя
7. **Игнорировать backlog** — статусы в `docs/04_Backlog.md` (BL-004 Skipped, BL-030 Blocked)

---

## Контакты / Escalation

TODO: Требуется уточнение (product owner, architect contacts)

При блокере → записать в `ai/OpenQuestions.md` с priority.
