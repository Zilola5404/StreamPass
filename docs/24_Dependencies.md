# StreamPass — Dependencies

> Дата: 2026-08-03

---

## Go (Backend)

**Module:** `streampass` | **Go version:** 1.22.2  
**Source:** `go.mod`

| Dependency | Version | Purpose | Type |
|------------|---------|---------|------|
| golang.org/x/crypto | v0.24.0 | Argon2id password hashing | Vendored (`vendor-src/crypto`) |
| golang.org/x/sys | v0.21.0 | Transitive (cpu for argon2) | Vendored (`vendor-src/sys`) |
| github.com/lib/pq | v1.10.9 | PostgreSQL driver | Vendored (`vendor-src/pq`) |

### Custom (dependency-free)

| Package | Location | Purpose |
|---------|----------|---------|
| JWT (HS256) | `backend/.../security/jwt_minimal.go` | Token signing/verification |
| Redis (RESP2) | `backend/.../redisclient/` | Session store |
| YAML parser | `shared/config/` | Config loading |
| HTTP router | stdlib `net/http` ServeMux | Routing (Go 1.22+) |

### Replace directives

```go
replace golang.org/x/crypto => ./vendor-src/crypto
replace golang.org/x/sys => ./vendor-src/sys
replace github.com/lib/pq => ./vendor-src/pq
```

---

## Flutter (Client)

**Source:** `client/pubspec.yaml` | **Version:** 0.1.0 | **Dart SDK:** >=3.3.0 <4.0.0

| Package | Version | Purpose |
|---------|---------|---------|
| flutter | sdk | UI framework |
| google_fonts | ^6.2.1 | Typography |
| http | ^1.2.1 | REST API client |
| shared_preferences | ^2.2.3 | Token/settings storage |
| provider | ^6.1.2 | State management |
| url_launcher | ^6.3.0 | Open payment URL |
| flutter_lints | ^4.0.0 (dev) | Linting |

---

## Go Core (Client Tunnel)

**Module:** `streampass/go_core` | **Go version:** 1.22.2  
**Source:** `client/go_core/go.mod`

| Dependency | Status |
|------------|--------|
| golang.org/x/mobile | Replace → `../vendor-src/mobile` — **NOT FOUND** |

---

## Docker Images

| Service | Image | Version |
|---------|-------|---------|
| PostgreSQL | postgres | 16-alpine |
| Redis | redis | 7-alpine |
| Caddy | caddy | 2-alpine |
| Backend | Built from `backend/Dockerfile` | Go 1.22-bookworm → distroless |
| Healthmonitor | Built from `backend/cmd/healthmonitor/Dockerfile` | Go 1.22 |

---

## External Services

| Service | Purpose | Status |
|---------|---------|--------|
| ЮKassa (yookassa.ru) | Payment processing | Client coded, not live-tested |
| Hysteria2 | VPN relay transport | External VPS, not in repo |
| nip.io | Dynamic DNS for MVP | `212-43-156-33.nip.io` |

---

## Missing / TODO

| Item | Impact |
|------|--------|
| `go.sum` | Non-reproducible builds |
| `vendor-src/mobile` | gomobile build fails |
| `streampasscore.aar` | Android tunnel bridge fails |
| CI/CD deps | No automated pipeline |

---

## Dependency Policy

1. Prefer stdlib and dependency-free implementations
2. New external deps require ADR (`docs/11_Decisions.md`)
3. Vendoring via `vendor-src/` with `replace` in go.mod
4. No `go get` without documenting in this file
5. Pin versions in go.mod / pubspec.yaml
