# AI Notes

> Working notes for AI agents. Not formal documentation.

---

## 2026-08-03: Initial Analysis

### Key Findings
- Real monorepo: Go backend + Flutter Android + Docker Compose
- Backend is mature (~80% MVP), well-structured Clean Architecture
- Client UI exists but VPN is stub — this is THE blocker
- README has outdated Health Monitor section (worker exists at cmd/healthmonitor)
- Old 03_CurrentState had fake data (OAuth, Telegram) — corrected
- No CI/CD, no go.sum, no integration tests
- Payments coded but never called live YooKassa API

### Dependency Strategy
- Most infra is dependency-free (JWT, Redis, YAML) — intentional
- vendor-src/ for crypto, sys, pq — due to sandbox restrictions during dev
- vendor-src/mobile MISSING — blocks gomobile

### Client API URL
- Default: `https://212-43-156-33.nip.io/api/v1` (compile-time dart-define)
- Set in `client/lib/main.dart`

### Relay Setup
- External Hysteria2 on VPS, not in docker-compose
- connection_config stored in relay_servers table (migration 0002)
- GET /servers requires JWT (returns secrets)

### Git
- 8 commits on main, all in Aug 2026
- docs/ai/prompts/scripts/templates — untracked (not committed yet)

### What I Did NOT Change
- Zero code changes — documentation only per task requirements

---

## Tips for Next AI Session

1. Start with BL-001 (tunnel) — everything else waits
2. Read `client/go_core/README.md` before touching go_core
3. Test VPN on real device, not just emulator
4. Don't trust README blindly — verify against code
5. `go test ./...` should be green before any backend changes

---

## Useful Commands

```bash
# Full stack
docker compose up -d --build && curl localhost:8080/health

# Backend dev
cd backend && go run ./cmd/server

# Client dev (emulator API)
cd client && flutter run --dart-define=STREAMPASS_API_URL=http://10.0.2.2:8080/api/v1

# Check routes
grep -r "HandleFunc\|Handle(" backend/internal/infrastructure/http/router/
```
