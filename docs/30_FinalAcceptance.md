# StreamPass — Final Acceptance Criteria

> Дата: 2026-08-11 | Based on ТЗ §22  
> Client under test: **v0.1.1+53** (device); +54 polish in tree  
> Sign-off pack: `reports/QA/MVP-signoff-package-2026-08-11.md`

---

## Functional Criteria (MVP)

| # | Criterion | Status | Evidence |
|---|-----------|--------|----------|
| F1 | User registration and authorization | ✅ | POST /register, /login, /logout, /refresh + Android onboarding |
| F2 | Payment and subscription activation | ⚠️ | Backend coded; ЮKassa live Blocked (BL-040 — no keys) |
| F3 | One-button connect | ✅ | Hysteria2 + TUN + AAR; APK **v0.1.1+50** |
| F4 | Automatic routing by rules | ✅ | Decision Engine + Rule Engine (BL-005, BL-006); ADR-018 |
| F5 | Foreign services via relay | ✅ | adb +50: ifconfig/ipify → NL; Gemini/LinkedIn RELAY |
| F6 | Russian services direct | ✅ | `*.ru` + PinDirect; 2ip DIRECT; gov app bypass |
| F7 | Auto relay switch on failure | ✅ | Fallback ports / relay selection (BL-017 + client logic) |
| F8 | Auto rule/config update | ✅ | Client polling + config auto-update (BL-006, BL-022) |

**Functional MVP: 7/8 fully done, 1/8 partial (payments live)**

---

## Technical Criteria (MVP)

| # | Criterion | Target | Status |
|---|-----------|--------|--------|
| T1 | Client startup time | ≤ 2s | ✅ Formal device pass (BL-053, 2026-08-10) |
| T2 | Connection time | ≤ 5s | ✅ Formal device pass (BL-053) |
| T3 | Auto-recovery | ≤ 10s | ⚠️ Retest 2026-08-11 attempted; formal stopwatch still open (see BL-053 retest) |
| T4 | Server availability | ≥ 99.9% | ⚠️ Not measured over 30d |
| T5 | Rule update without reinstall | Required | ✅ API versioning + client polling |

**Technical MVP: 3/5 done; T3 partial, T4 ops pending**

---

## Quality Criteria

| # | Criterion | Status |
|---|-----------|--------|
| Q1 | All unit tests pass | ✅ `go test ./...` green |
| Q2 | Flutter tests pass | ✅ `flutter test` green |
| Q3 | No critical security issues | ⚠️ Checklist verified 2026-08-10; pre-prod items remain |
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

1. **ЮKassa live-tested** (BL-040 — needs live keys)
2. **Instagram feed quality** (BUG-IG-FEED — QUIC/media throughput; routing OK)
3. **T3 recover formal** + **T4 30d uptime** (ops)
4. Human sign-off (PO / TL / QA)

VPN tunnel + Decision Engine + split routing recheck on **+50** are **no longer** blockers.

---

## Beta Acceptance (Additional)

- [x] CI/CD green on every push
- [x] Integration tests for auth, billing, relay
- [x] Production Android signing path (key.properties)
- [ ] Real domain with HTTPS (nip.io works for MVP)
- [ ] 10+ beta users successfully connected
- [x] No open critical tunnel bugs in `docs/05_Bugs.md`
- [x] Security checklist MVP pass (`docs/28_SecurityChecklist.md`; report `reports/Security/28-SecurityChecklist-verification-2026-08-10.md`; pre-production items remain)

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

**Connect + split-routing MVP is functionally ready on +50; formal acceptance not complete.**  
Next queue: Instagram throughput A/B (adb) → BL-040 keys → sign-off.  
BL-054 Terms in-app Done (+51).  
Backend API + Admin + monitoring + backups: operational on prod.
