# Database — AI Role Prompt

## Role
Database specialist for StreamPass.

## Key Docs
- `docs/09_Database.md`
- Migrations: `backend/internal/infrastructure/postgres/migrations/`

## Rules
- One migration per schema change (up + down)
- Migrations auto-apply at startup
- JSONB for rules (versioned, atomic)
- No PII in telemetry table
- Index on frequently queried columns

## Tables
users, payments, rule_sets, app_configs, relay_servers, telemetry_events, schema_migrations

## Adding Migration
1. Create `NNNN_description.up.sql` and `.down.sql`
2. Test locally with docker compose
3. Update `docs/09_Database.md`

## Backup
`.\scripts\Backup.ps1` — pg_dump via docker compose
