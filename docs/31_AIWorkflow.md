# StreamPass — AI Workflow

> Дата: 2026-08-03

---

## Workflow Overview

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Read Docs  │────▶│  Pick Task   │────▶│  Implement  │
│  + ai/      │     │  from Backlog│     │  + Test     │
└─────────────┘     └──────────────┘     └──────┬──────┘
                                                 │
┌─────────────┐     ┌──────────────┐     ┌──────▼──────┐
│  Handoff    │◀────│  Update Docs │◀────│  Smoke Test │
│  LastSession│     │  + Progress  │     │             │
└─────────────┘     └──────────────┘     └─────────────┘
```

---

## Session Start Protocol

1. Read `docs/00_ProjectRules.md`
2. Read `docs/14_AIContext.md`
3. Read `ai/CurrentTask.md` + `ai/LastSession.md`
4. Run `git log --oneline -5` + `git status`
5. Run `go build ./...` + `go test ./...`
6. Confirm task scope with user if ambiguous

---

## During Work

| Action | Rule |
|--------|------|
| Code change | Follow Clean Architecture, minimal diff |
| New API | Update `docs/08_API.md` |
| New migration | Update `docs/09_Database.md` |
| Architecture change | Write ADR in `docs/11_Decisions.md` |
| Bug found | Add to `docs/05_Bugs.md` |
| Blocker | Add to `ai/OpenQuestions.md` |
| Unsure | Don't change code, ask user |

---

## Session End Protocol

1. Run tests (`go test ./...`, `flutter test` if client changed)
2. Update `ai/LastSession.md` (what done, files changed, remaining)
3. Update `docs/10_Progress.md` (if task completed)
4. Update `docs/04_Backlog.md` (task status)
5. Update `ai/CurrentTask.md` + `ai/NextTask.md`
6. Commit only if user explicitly requested

---

## Task Types & Doc Updates

| Task Type | Required Updates |
|-----------|-----------------|
| Backend feature | 08_API, 10_Progress, 04_Backlog, ai/LastSession |
| Client feature | 03_CurrentState, 10_Progress, ai/LastSession |
| Bug fix | 05_Bugs (status→Fixed), 10_Progress |
| Architecture | 07_Architecture, 11_Decisions |
| Documentation only | 10_Progress, ai/LastSession |
| Infrastructure | 26_Deployment, 25_Environment |

---

## AI Role Selection

Use role-specific prompts from `prompts/`:

| Task | Prompt File |
|------|-------------|
| Architecture design | `prompts/SeniorArchitect.md` |
| Backend coding | `prompts/BackendDeveloper.md` |
| Flutter UI | `prompts/FrontendDeveloper.md` |
| Android VPN/native | `prompts/MobileDeveloper.md` |
| Security review | `prompts/SecurityEngineer.md` |
| Code review | `prompts/CodeReviewer.md` |
| Test writing | `prompts/QAEngineer.md` |
| Docker/CI/deploy | `prompts/DevOpsEngineer.md` |

Also available: `prompts/00_SystemPrompt.md` through `prompts/13_API.md`

---

## Quality Gates

Before marking task Done (`docs/15_DefinitionOfDone.md`):

```
✓ go build ./...
✓ go vet ./...
✓ go test ./...
✓ flutter analyze (if client)
✓ flutter test (if client)
✓ docs updated
✓ no secrets in diff
✓ no unrelated changes
```

---

## Multitasking Rules

- One task at a time (Project Rule #6)
- Finish current task before starting next
- Update `ai/CurrentFocus.md` if focus shifts
- Don't leave partial work without LastSession update

---

## Handoff Between AI Tools

Works across: Cursor, Claude Code, GitHub Copilot, Codex.

All tools read same docs:
- `docs/14_AIContext.md` — project context
- `docs/16_AI_HANDOFF.md` — handoff protocol
- `ai/LastSession.md` — previous session state

No tool-specific state files. All state in git-tracked docs.
