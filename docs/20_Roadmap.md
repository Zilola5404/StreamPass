# StreamPass — Roadmap

> Дата: 2026-08-03 | Based on ТЗ §23-24 and current state

---

## Phase 0: Documentation (DONE)

- [x] AI-friendly documentation structure
- [x] Project analysis and reports
- [x] Role-specific AI prompts

---

## Phase 1: MVP Completion (CURRENT)

**Goal:** End-to-end VPN connect on Android with backend.

| Week | Tasks | Status |
|------|-------|--------|
| W1 | Hysteria2 tunnel in go_core, build AAR | ❌ |
| W1 | Wire AAR to Android VPNService | ❌ |
| W2 | Decision Engine + Rule Engine on client | ❌ |
| W2 | Live-test ЮKassa sandbox | ❌ |
| W3 | E2E test: register → pay → connect → verify IP | ❌ |
| W3 | CI/CD GitHub Actions | ❌ |

**Exit criteria:** User can connect with one button, foreign IP verified.

---

## Phase 2: Beta (Q4 2026 — TODO dates)

- Integration tests (testcontainers)
- Production Android signing
- Real domain + HTTPS
- Error monitoring
- Load testing
- Closed beta (10-50 users)

---

## Phase 3: Production v1.0

- App Store / Google Play release
- Prometheus + Grafana
- Backup automation
- Security audit
- Performance benchmarks (ТЗ §22)

---

## Phase 4: Multi-Platform (post-v1.0)

| Platform | Adapter | Priority |
|----------|---------|----------|
| Windows 10/11 | WFP | P1 |
| iOS 17+ | Network Extension | P1 |
| macOS 13+ | Network Extension | P2 |

Shared Go core target: 90% code reuse (ТЗ §4).

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
| VPN Tunnel | — | ❌ Blocker |
| Beta Launch | TODO | ❌ |
| Production v1.0 | TODO | ❌ |
