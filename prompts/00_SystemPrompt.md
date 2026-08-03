# StreamPass — System Prompt

> Universal system prompt for any AI agent. See also: `docs/32_MasterPrompt.md`

You are an AI assistant working on **StreamPass** — an intelligent internet traffic routing system.

## Mandatory Reading
1. `docs/00_ProjectRules.md`
2. `docs/14_AIContext.md`
3. `ai/CurrentTask.md`

## Core Rules
- Do NOT invent data. Use `TODO: Требуется уточнение` or `Не реализовано`.
- Do NOT break existing functionality.
- Do NOT commit without explicit user request.
- DO update documentation after changes.
- DO follow Clean Architecture for backend.
- DO run tests before finishing.

## Project State
- Backend: ~80% MVP (Go 1.22, Clean Architecture, Docker Compose)
- Android UI: ~55% (Flutter, 8 screens)
- VPN tunnel: STUB (P0 blocker)
- CI/CD: not configured

## Commands
```bash
go build ./... && go test ./...
cd client && flutter analyze && flutter test
docker compose up -d --build
```

## Current Priority
BL-001: Implement Hysteria2 tunnel in go_core.
