# StreamPass — Final Acceptance Criteria

> Дата: 2026-08-03 | Based on ТЗ §22

---

## Functional Criteria (MVP)

| # | Criterion | Status | Evidence |
|---|-----------|--------|----------|
| F1 | User registration and authorization | ✅ | POST /register, /login, /logout + Android onboarding |
| F2 | Payment and subscription activation | ⚠️ | Backend coded, ЮKassa not live-tested |
| F3 | One-button connect | ❌ | UI exists, VPN tunnel stub |
| F4 | Automatic routing by rules | ❌ | Decision Engine not on client |
| F5 | Foreign services via relay | ❌ | Depends on F3 |
| F6 | Russian services direct | ❌ | Depends on F4 |
| F7 | Auto relay switch on failure | ❌ | Client-side logic not implemented |
| F8 | Auto rule/config update | ⚠️ | API exists, client polling not verified |

**Functional MVP: 1/8 fully done, 2/8 partial**

---

## Technical Criteria (MVP)

| # | Criterion | Target | Status |
|---|-----------|--------|--------|
| T1 | Client startup time | ≤ 2s | TODO: Not measured |
| T2 | Connection time | ≤ 5s | ❌ Tunnel stub |
| T3 | Auto-recovery | ≤ 10s | ❌ Not implemented |
| T4 | Server availability | ≥ 99.9% | TODO: Not measured |
| T5 | Rule update without reinstall | Required | ✅ API versioning |

**Technical MVP: 1/5 done, 0/5 measured**

---

## Quality Criteria

| # | Criterion | Status |
|---|-----------|--------|
| Q1 | All unit tests pass | ✅ `go test ./...` green |
| Q2 | Flutter tests pass | ✅ `flutter test` green |
| Q3 | No critical security issues | ⚠️ Debug signing, no live security audit |
| Q4 | Documentation complete | ✅ (this initialization) |
| Q5 | CI/CD pipeline | ❌ Not configured |
| Q6 | Integration tests | ❌ Not implemented |
| Q7 | Docker deploy works | ✅ docker-compose.yml |

---

## Acceptance Sign-Off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Product Owner | TODO | — | — |
| Tech Lead | TODO | — | — |
| QA | TODO | — | — |

---

## MVP Acceptance Blockers

1. **VPN tunnel must work end-to-end** (BL-001, BL-002, BL-003)
2. **Decision Engine on client** (BL-005, BL-006)
3. **ЮKassa live-tested** (BL-004)
4. **Performance targets measured and met** (T1-T4)

---

## Beta Acceptance (Additional)

- [ ] CI/CD green on every push
- [ ] Integration tests for auth, billing, relay
- [ ] Production Android signing
- [ ] Real domain with HTTPS
- [ ] 10+ beta users successfully connected
- [ ] No critical bugs in `docs/05_Bugs.md`
- [ ] Security checklist passed (`docs/28_SecurityChecklist.md`)

---

## Production Acceptance (Additional)

- [ ] All Beta criteria met
- [ ] App Store / Google Play approved
- [ ] Monitoring (Prometheus/Grafana) operational
- [ ] Backup/restore tested
- [ ] Load test passed (50 RPS, p99 < 500ms)
- [ ] Privacy policy published
- [ ] 99.9% uptime over 30 days

---

## Verdict

**MVP NOT READY for acceptance.**  
Primary blocker: VPN tunnel (go_core stub).  
Backend API: ready for integration testing.  
Android UI: ready, pending tunnel.
