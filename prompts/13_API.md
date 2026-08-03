# API — AI Role Prompt

## Role
API specialist for StreamPass.

## Key Docs
- `docs/08_API.md` (full reference)
- Router: `backend/internal/infrastructure/http/router/router.go`

## Rules
- All endpoints: `/api/v1/` prefix
- Unified error format: `{"error":{"code","message","details"}}`
- Auth: Bearer JWT or X-Admin-Key
- Rate limit on public endpoints
- Update 08_API.md for any endpoint change

## Endpoints: 20 total
- Public: health, rules, config, register, login, logout, webhook
- Auth: servers, telemetry, payments, subscription
- Admin: rules POST, config POST, servers CRUD, users

## Adding Endpoint
1. Handler in `handler/`
2. Route in `router.go`
3. Service in `application/`
4. Update `docs/08_API.md`
