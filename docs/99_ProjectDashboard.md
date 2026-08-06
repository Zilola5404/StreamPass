# StreamPass — Project Dashboard

> Last updated: 2026-08-06

---

## Status at a Glance

| Area | Progress | Status |
|------|----------|--------|
| Backend API | 95% | 🟢 Functional (prod) |
| Android UI + VPN | 90% | 🟢 Connect + rules + regions |
| VPN Tunnel | 90% | 🟢 Hysteria2 (not stub) |
| Billing | 60% | 🟡 Code ready, live Skipped (BL-004) |
| CI/CD | 100% | 🟢 GitHub Actions |
| Admin Panel | 100% | 🟢 `/admin/` |
| Monitoring / Backup | 100% | 🟢 Grafana/Prometheus + daily backups |
| Documentation | 100% | 🟢 Kept current |
| Tests (unit) | 60% | 🟢 Partial + growing |
| Tests (integration) | 80% | 🟢 BL-011 + SmokeTest + loadtest |

**Overall MVP: ~85%**

---

## Active Sprint

**Goal:** Stabilize Android MVP on prod; optional monetization / multi-OS only on request  
**Focus:** Device manual connect +25; off-site backup verify; YooKassa only if explicitly requested

See: `ai/CurrentSprint.md`

---

## Top Blockers

1. 🟡 ЮKassa not live-tested — BL-004 **Skipped** (intentional)
2. 🟡 BL-030 auto-renewal **Blocked** on BL-004
3. ⚪ BL-023/024/025 Windows / iOS / macOS — **Open intentional** (do not start without request)
4. ⚪ Measured client perf (T1–T4) + physical device recheck

---

## Recent Activity

| Date | Event |
|------|-------|
| 2026-08-05 | Docs sync to product reality (VPN Done, admin, monitoring, +17) |
| 2026-08-04 | BL-020…022,026,027,031–033; APK v0.1.1+17 |
| 2026-08-04 | CI, signing, DNS/DoH, exclusions, Decision Engine |
| 2026-08-03 | BL-001…003 VPN tunnel + AAR + E2E verify |
| 2026-08-03 | AI documentation initialized |

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
| Prod | `https://212-43-156-33.nip.io` |
| Admin | `/admin/` |
| APK | v0.1.1+25 |
| Relays (prod) | NL nodes only (multi-region software ready) |
| Open bugs (active) | 0 critical tunnel bugs |
| Open backlog intentional | BL-023…025, BL-004 Skipped, BL-030 Blocked |

---

## Next Milestone

**MVP acceptance polish** — device E2E recheck; optional ЮKassa live; measured T1–T4.  
Not blockers for connect path: BL-001…003,005,006 Done.
