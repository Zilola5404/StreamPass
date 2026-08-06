# StreamPass — Final Acceptance Criteria

> Дата: 2026-08-05 | Based on ТЗ §22

---

## Functional Criteria (MVP)

| # | Criterion | Status | Evidence |
|---|-----------|--------|----------|
| F1 | User registration and authorization | ✅ | POST /register, /login, /logout, /refresh + Android onboarding |
| F2 | Payment and subscription activation | ⚠️ | Backend coded; ЮKassa live Skipped (BL-004) |
| F3 | One-button connect | ✅ | Hysteria2 + TUN + AAR (BL-001…003); APK v0.1.1+25 |
| F4 | Automatic routing by rules | ✅ | Decision Engine + Rule Engine (BL-005, BL-006) |
| F5 | Foreign services via relay | ✅ | Depends on F3 — verified on BL-003 path |
| F6 | Russian services direct | ✅ | DIRECT rules via Decision Engine |
| F7 | Auto relay switch on failure | ✅ | Fallback ports / relay selection (BL-017 + client logic) |
| F8 | Auto rule/config update | ✅ | Client polling + config auto-update (BL-006, BL-022) |

**Functional MVP: 7/8 fully done, 1/8 partial (payments live)**

---

## Technical Criteria (MVP)

| # | Criterion | Target | Status |
|---|-----------|--------|--------|
| T1 | Client startup time | ≤ 2s | ⚠️ Not measured on device |
| T2 | Connection time | ≤ 5s | ⚠️ Path works; not formally measured |
| T3 | Auto-recovery | ≤ 10s | ⚠️ Logic present; not formally measured |
| T4 | Server availability | ≥ 99.9% | ⚠️ Not measured over 30d |
| T5 | Rule update without reinstall | Required | ✅ API versioning + client polling |

**Technical MVP: 1/5 done, 0/4 measured (T1–T4 remain)**

---

## Quality Criteria

| # | Criterion | Status |
|---|-----------|--------|
| Q1 | All unit tests pass | ✅ `go test ./...` green |
| Q2 | Flutter tests pass | ✅ `flutter test` green |
| Q3 | No critical security issues | ⚠️ No live security audit; release signing path Done (BL-013) |
| Q4 | Documentation complete | ✅ |
| Q5 | CI/CD pipeline | ✅ BL-010 |
| Q6 | Integration tests | ✅ BL-011 + SmokeTest |
| Q7 | Docker deploy works | ✅ docker-compose.yml (prod) |

---

## Acceptance Sign-Off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Product Owner | TODO | — | — |
| Tech Lead | TODO | — | — |
| QA | TODO | — | — |

---

## MVP Acceptance Blockers (remaining)

1. **ЮKassa live-tested** (BL-004 Skipped intentional)
2. **Performance targets measured and met** (T1–T4)
3. **Physical device recheck** APK v0.1.1+25

VPN tunnel + Decision Engine are **no longer** blockers (BL-001…003,005,006 Done).

---

## Beta Acceptance (Additional)

- [x] CI/CD green on every push
- [x] Integration tests for auth, billing, relay
- [x] Production Android signing path (key.properties)
- [ ] Real domain with HTTPS (nip.io works for MVP)
- [ ] 10+ beta users successfully connected
- [x] No open critical tunnel bugs in `docs/05_Bugs.md`
- [ ] Security checklist fully passed (`docs/28_SecurityChecklist.md`)

---

## Production Acceptance (Additional)

- [ ] All Beta criteria met
- [ ] App Store / Google Play approved
- [x] Monitoring (Prometheus/Grafana) operational (local)
- [x] Backup/restore path (daily cron; off-site optional)
- [x] Load test baseline passed (BL-032; expand as needed)
- [ ] Privacy policy published
- [ ] 99.9% uptime over 30 days

---

## Verdict

**Connect path MVP is functionally ready; formal acceptance not complete.**  
Remaining: ЮKassa live, measured T1–T4, device recheck.  
Backend API + Admin + monitoring + backups: operational on prod.
