# StreamPass — Coding Standards

> Дата: 2026-08-03

---

## Go (Backend + Shared)

### Architecture
- Clean Architecture: `domain` → `application` → `infrastructure`
- Handlers: decode request → call service → encode response. **No business logic in handlers.**
- DI via constructor functions, wiring in `main.go`
- Domain models in `backend/internal/domain/<context>/`
- Use cases in `backend/internal/application/<context>/`

### Style
- Go 1.22+, follow effective Go
- Package comments on every package
- Exported functions/types documented
- Errors: `shared/errors.AppError` with codes
- Context passed as first parameter
- No `panic` in production code (middleware.Recover catches)

### Naming
- Packages: lowercase, single word (`auth`, `relay`, `configsvc`)
- Interfaces: verb-based (`PaymentProvider`, `TokenVerifier`)
- Handlers: `<Domain>Handler` with methods matching HTTP verbs
- Repositories: `<Entity>Repository` in postgres package

### Testing
- Table-driven tests preferred
- `_test.go` in same package
- Mock external deps (Redis: mock server in client_test.go)
- Run: `go test ./...`, `go vet ./...`

### Dependencies
- Prefer stdlib
- New external deps require ADR in `docs/11_Decisions.md`
- Vendoring via `vendor-src/` with `replace` in go.mod

---

## Dart/Flutter (Client)

### Architecture
- Screens in `lib/screens/`
- Services in `lib/services/`
- Provider for state management
- Native bridge via MethodChannel/EventChannel

### Style
- Dart >=3.3.0, follow effective dart
- `flutter_lints` rules
- Widgets: prefer StatelessWidget where possible
- Async: use async/await, handle errors

### Testing
- Unit tests in `test/`
- Widget tests for key screens
- Run: `flutter analyze`, `flutter test`

---

## Kotlin (Android Native)

### Style
- Kotlin idioms, JVM 17
- VPNService follows Android VPN API guidelines
- MethodChannel names match Dart side exactly

---

## General

- No secrets in code — env vars only
- No PII in logs or telemetry
- Minimal diff — don't refactor unrelated code
- Comments explain *why*, not *what*
- API changes require `docs/08_API.md` update
