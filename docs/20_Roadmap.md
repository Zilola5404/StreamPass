# StreamPass — Roadmap

> Дата: 2026-08-05 | Based on ТЗ §23-24 and current state

---

## Phase 0: Documentation (DONE)

- [x] AI-friendly documentation structure
- [x] Project analysis and reports
- [x] Role-specific AI prompts

---

## Phase 1: MVP Completion (MOSTLY DONE)

**Goal:** End-to-end VPN connect on Android with backend.

| Week | Tasks | Status |
|------|-------|--------|
| W1 | Hysteria2 tunnel in go_core, build AAR | ✅ Done (BL-001, BL-002) |
| W1 | Wire AAR to Android VPNService | ✅ Done (BL-003) |
| W2 | Decision Engine + Rule Engine on client | ✅ Done (BL-005, BL-006) |
| W2 | Live-test ЮKassa sandbox | ⏭️ Skipped (BL-004) |
| W3 | E2E path: register → connect → verify IP | ✅ Done (BL-003); pay live skipped |
| W3 | CI/CD GitHub Actions | ✅ Done (BL-010) |

**Exit criteria:** User can connect with one button, foreign IP verified — **met** (payments live still open).

**Remaining Phase 1 polish:** device manual connect +25; optional ЮKassa live on request.

---

## Phase 2: Beta (Q4 2026 — TODO dates)

- [x] Integration tests (BL-011)
- [x] Production Android signing path (BL-013)
- [ ] Real domain + HTTPS (nip.io OK for now)
- [x] Error / ops monitoring baseline (BL-021 Grafana/Prometheus)
- [x] Load testing baseline (BL-032)
- [ ] Closed beta (10-50 users)

---

## Phase 3: Production v1.0

- [ ] App Store / Google Play release
- [x] Prometheus + Grafana (BL-021)
- [x] Backup automation (BL-033 daily; off-site optional)
- [ ] Security audit
- [ ] Performance benchmarks measured (ТЗ §22 T1–T4)

---

## Phase 4: Multi-Platform (post-v1.0)

| Platform | Adapter | Priority | Status |
|----------|---------|----------|--------|
| Windows 10/11 | WFP | P1 | Open BL-023 |
| iOS 17+ | Network Extension | P1 | Open BL-024 |
| macOS 13+ | Network Extension | P2 | Open BL-025 |

Shared Go core target: 90% code reuse (ТЗ §4). Do not start without explicit request.

---

## Phase 5: Growth (v1.1 — v2.0)

From ТЗ §24:

| Version | Features |
|---------|----------|
| v1.1 | Linux, additional relays, improved Decision Engine |
| v1.2 | TUIC transport, advanced routing |
| v1.5 | Score Engine (telemetry-based, no ML) |
| v2.0 | Custom transport, Multipath QUIC, adaptive routing |

---

## Not Planned (MVP scope exclusion)

Kubernetes, ML/AI routing, MASQUE, ASN/GeoIP, browser extension, corporate version, multi-hop, OpenWRT, referral system.

---

## Milestone Tracking

| Milestone | Target | Status |
|-----------|--------|--------|
| Backend MVP | Done | ✅ |
| Android UI | Done | ✅ |
| VPN Tunnel | Done | ✅ |
| Admin / Monitoring / Backup | Done | ✅ |
| Beta Launch | TODO | ❌ |
| Production v1.0 | TODO | ❌ |
