# StreamPass — Project Dashboard

> Last updated: 2026-08-03

---

## Status at a Glance

| Area | Progress | Status |
|------|----------|--------|
| Backend API | 80% | 🟢 Functional |
| Android UI | 55% | 🟡 UI ready, VPN stub |
| VPN Tunnel | 5% | 🔴 Stub only |
| Billing | 60% | 🟡 Code ready, not live-tested |
| CI/CD | 0% | 🔴 Not configured |
| Documentation | 100% | 🟢 Initialized |
| Tests (unit) | 40% | 🟡 Partial coverage |
| Tests (integration) | 0% | 🔴 Not implemented |

**Overall MVP: ~35%**

---

## Active Sprint

**Goal:** Documentation initialization + identify MVP blockers  
**Focus:** VPN tunnel implementation (next)

See: `ai/CurrentSprint.md`

---

## Top Blockers

1. 🔴 VPN tunnel stub (go_core) — BL-001
2. 🔴 streampasscore.aar not built — BL-002
3. 🟡 ЮKassa not live-tested — BL-004
4. 🟡 No CI/CD — BL-010

---

## Recent Activity

| Date | Event |
|------|-------|
| 2026-08-03 | AI documentation initialized |
| — | Backend linked with Android client (6a6f029) |
| — | Relay data fix in VPN service (50e9e15) |
| — | First MVP release (9570b95) |

---

## Quick Links

| Document | Purpose |
|----------|---------|
| [AI Context](14_AIContext.md) | Start here (AI) |
| [Current State](03_CurrentState.md) | What works |
| [Backlog](04_Backlog.md) | Task list |
| [API](08_API.md) | API reference |
| [Architecture](07_Architecture.md) | System design |
| [Bugs](05_Bugs.md) | Known bugs |
| [Risks](13_Risks.md) | Risk register |

---

## Metrics

| Metric | Value |
|--------|-------|
| Git commits | 8 |
| Backend Go files | ~50+ |
| Flutter screens | 8 |
| API endpoints | 20 |
| DB tables | 7 |
| Unit tests (Go) | 7 files |
| Unit tests (Flutter) | 3 files |
| Open bugs | 5 |
| Open backlog items | 20+ |

---

## Next Milestone

**MVP VPN Working** — user can connect with one button, foreign IP verified.  
Blockers: BL-001, BL-002, BL-003, BL-005, BL-006
