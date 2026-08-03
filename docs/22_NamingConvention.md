# StreamPass — Naming Conventions

> Дата: 2026-08-03

---

## Repository Structure

| Path | Convention | Example |
|------|------------|---------|
| `backend/internal/domain/` | Bounded context name | `user/`, `relay/`, `rule/` |
| `backend/internal/application/` | `<context>/service.go` | `auth/service.go` |
| `backend/internal/infrastructure/` | Technology layer | `postgres/`, `redisclient/`, `http/` |
| `shared/` | Cross-cutting packages | `config/`, `errors/`, `logger/` |
| `client/lib/screens/` | `<name>_screen.dart` | `home_screen.dart` |
| `client/lib/services/` | `<name>_service.dart` | `auth_service.dart` |

---

## Go

| Element | Convention | Example |
|---------|------------|---------|
| Package | lowercase, no underscores | `configsvc`, `redisclient` |
| File | snake_case rare; usually single word | `service.go`, `user_repository.go` |
| Type | PascalCase | `AuthHandler`, `TokenPair` |
| Interface | PascalCase, often `-er` suffix | `PaymentProvider`, `TokenVerifier` |
| Function (exported) | PascalCase | `NewAuthHandler`, `Execute` |
| Function (unexported) | camelCase | `writeTokenPair`, `statusForCode` |
| Constant | PascalCase or camelCase | `apiV1Prefix`, `TimeFormat` |
| Error codes | snake_case string | `"invalid_credentials"`, `"rate_limited"` |
| DB table | snake_case plural | `users`, `relay_servers`, `rule_sets` |
| DB column | snake_case | `subscription_active_until`, `connection_config` |
| Migration | `NNNN_description.up.sql` | `0001_init.up.sql` |
| Env var | UPPER_SNAKE_CASE | `JWT_SECRET`, `ADMIN_API_KEY` |
| Config YAML key | snake_case | `http_port`, `max_open_conns` |

---

## API

| Element | Convention | Example |
|---------|------------|---------|
| Base path | `/api/v1/` | All business endpoints |
| Route pattern | `"METHOD /path"` | `"POST /register"` |
| JSON field | snake_case | `access_token`, `rtt_ms` |
| Error code | snake_case | `invalid_input`, `token_expired` |
| Admin header | `X-Admin-Key` | Static API key |
| Auth header | `Authorization: Bearer <token>` | JWT access token |

---

## Flutter/Dart

| Element | Convention | Example |
|---------|------------|---------|
| File | snake_case | `auth_service.dart` |
| Class | PascalCase | `AuthService`, `HomeScreen` |
| Variable | camelCase | `accessToken`, `isConnected` |
| Constant | camelCase or lowerCamelCase | `defaultApiUrl` |
| Private | `_prefix` | `_token`, `_loadSettings()` |

---

## Docker

| Element | Convention | Example |
|---------|------------|---------|
| Service name | lowercase | `postgres`, `backend`, `healthmonitor` |
| Volume | `<service>_data` | `postgres_data`, `caddy_data` |
| Image tag | version or alpine | `postgres:16-alpine` |

---

## Documentation

| Element | Convention | Example |
|---------|------------|---------|
| Doc file | `NN_Name.md` | `08_API.md`, `14_AIContext.md` |
| Backlog ID | `BL-NNN` | `BL-001` |
| Bug ID | `BUG-NNN` | `BUG-001` |
| ADR ID | `ADR-NNN` | `ADR-001` |
| Tech Debt ID | `TD-NNN` | `TD-001` |
| Risk ID | `TR/BR/IR/SR-NN` | `TR-01` |

---

## Git

| Element | Convention | Example |
|---------|------------|---------|
| Branch | `feature/description` or `fix/description` | `feature/hysteria2-tunnel` |
| Commit | Imperative, concise | `fix: restore relay data in VPN service` |
