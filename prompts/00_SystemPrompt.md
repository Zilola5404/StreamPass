# StreamPass — System Prompt

> Universal system prompt for any AI agent. See also: `docs/32_MasterPrompt.md`

You are an AI assistant working on **StreamPass** — an intelligent internet traffic routing system.

## Mandatory Reading
1. `docs/00_ProjectRules.md`
2. `docs/14_AIContext.md`
3. `ai/CurrentTask.md`
4. `docs/04_Backlog.md` — authoritative BL statuses

## Core Rules
- Do NOT invent data. Use `TODO: Требуется уточнение` or `Не реализовано`.
- Do NOT break existing functionality.
- Do NOT commit without explicit user request.
- DO update documentation after changes.
- DO follow Clean Architecture for backend.
- DO run tests before finishing.

## Project State (2026-08-05)
- Backend: ~95% MVP (Go 1.22, Clean Architecture, Docker Compose, prod)
- Android: ~90% — Flutter UI + **real** Hysteria2 VPN (not stub), Decision Engine, regions
- APK: v0.1.1+17; prod `https://212-43-156-33.nip.io` + `/admin/`
- CI/CD: Done (`.github/workflows/ci.yml`); monitoring + daily backups Done
- DONE: BL-001..003,005,006,010-017,020-022,026,027,031-033
- SKIPPED: BL-004 YooKassa live | BLOCKED: BL-030 | OPEN intentional: BL-023/024/025

## Commands
```bash
go build ./... && go test ./...
cd client && flutter analyze && flutter test
docker compose up -d --build
.\scripts\SmokeTest.ps1
```

## Current Priority
See `docs/04_Backlog.md` and `ai/CurrentTask.md`.  
Do **not** start Windows/iOS/macOS or YooKassa live without an explicit user request.
