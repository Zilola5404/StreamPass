# StreamPass — Master Prompt

> Универсальный промпт для любого AI-агента, работающего с проектом StreamPass.  
> Дата: 2026-08-03

---

## System Context

You are working on **StreamPass** — an intelligent internet traffic routing system (MVP).  
Repository: Go 1.22 backend + Flutter Android client + Docker Compose infrastructure.

**Current stage:** Backend API functional, Android UI ready, VPN tunnel (Hysteria2) is a STUB.

---

## Before Any Task

1. Read `docs/00_ProjectRules.md` — project constitution
2. Read `docs/14_AIContext.md` — full AI context
3. Read `docs/03_CurrentState.md` — what's implemented
4. Check `ai/CurrentTask.md` — current priority
5. Check `ai/LastSession.md` — previous session state

---

## Rules

- **Do NOT invent data.** Write `TODO: Требуется уточнение` or `Не реализовано` when unsure.
- **Do NOT break existing functionality.** Run tests before and after.
- **Do NOT commit without explicit user request.**
- **Do NOT add features outside task scope.**
- **DO update documentation** after changes (Progress, API, Architecture as needed).
- **DO follow Clean Architecture** for backend changes.
- **DO use minimal diffs** — no drive-by refactoring.

---

## Architecture Summary

```
Flutter Android → HTTPS → Caddy → Go Backend → PostgreSQL + Redis
                                      ↕
                              Health Monitor → Relay VPS (Hysteria2)
```

- Backend: Clean Architecture (domain → application → infrastructure → http)
- API: `/api/v1/*`, JWT auth, X-Admin-Key for admin
- Client: Flutter UI + Android VPNService + go_core (stub)

---

## Key Commands

```bash
go build ./...                    # Build backend
go test ./...                     # Test backend
go vet ./...                      # Lint backend
cd client && flutter analyze      # Lint client
cd client && flutter test         # Test client
docker compose up -d --build      # Full stack
curl http://localhost:8080/health # Health check
```

---

## Current Priority (P0)

1. Implement Hysteria2 tunnel in go_core (BL-001)
2. Build streampasscore.aar (BL-002)
3. End-to-end VPN on Android (BL-003)

---

## Response Format

When completing a task, report:
1. What was done
2. Files changed
3. Test results
4. Documentation updated
5. Remaining work / blockers
6. Suggested next task

---

## Documentation Map

| Need | Read |
|------|------|
| API endpoints | `docs/08_API.md` |
| Database schema | `docs/09_Database.md` |
| Architecture | `docs/07_Architecture.md` |
| Task list | `docs/04_Backlog.md` |
| Product spec | `docs/02_TZ.md` |
| Bugs | `docs/05_Bugs.md` |
| Environment setup | `docs/25_Environment.md` |
| Deployment | `docs/26_Deployment.md` |
