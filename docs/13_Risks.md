# StreamPass — Риски

> Дата: 2026-08-05

---

## Технические риски

| ID | Риск | Вероятность | Impact | Митигация |
|----|------|-------------|--------|-----------|
| TR-01 | VPN tunnel (go_core) не реализован — MVP заблокирован | Low | Mitigated | Done BL-001…003 — Hysteria2 + AAR + E2E |
| TR-02 | Hysteria2 dependency — breaking changes в upstream | Medium | High | Pin version, monitor releases |
| TR-03 | Custom JWT/Redis — bugs in security-critical code | Medium | Critical | Security review, unit + integration tests |
| TR-04 | No go.sum — non-reproducible builds | Low | Mitigated | Done BL-027 — go.sum in repo |
| TR-05 | No integration tests — regressions undetected | Low | Mitigated | Done BL-011 + CI (BL-010) |
| TR-06 | vendor-src/mobile missing — gomobile build fails | Low | Mitigated | vendor-src/mobile present |

---

## Бизнес риски

| ID | Риск | Вероятность | Impact | Митигация |
|----|------|-------------|--------|-----------|
| BR-01 | ЮKassa не протестирована — payments fail in production | High | Critical | BL-004 Skipped intentional; live-test before monetization |
| BR-02 | Subscription cancel semantics unclear | Medium | Medium | Product decision (ADR-010) |
| BR-03 | Обещание «ускорение интернета» — user disappointment | Medium | High | Positioning: «стабильный маршрут» (ТЗ CTO note) |
| BR-04 | Только Android — limited market | Medium | Medium | Roadmap Open: BL-023 Windows, BL-024 iOS, BL-025 macOS |

---

## Инфраструктурные риски

| ID | Риск | Вероятность | Impact | Митигация |
|----|------|-------------|--------|-----------|
| IR-01 | Single VPS — no HA | High | Critical | Multi-relay software ready; future HA (post-MVP) |
| IR-02 | nip.io domain — not production-grade | High | Medium | Real domain before public production |
| IR-03 | No automated backup | Low | Mitigated | Done BL-033 daily cron; off-site copy still optional |
| IR-04 | No CI/CD — manual deploy errors | Low | Mitigated | Done BL-010 `.github/workflows/ci.yml` |
| IR-05 | Redis no persistence (no save/appendonly) | Medium | Medium | Sessions lost on restart — acceptable for MVP |
| IR-06 | No monitoring (Prometheus/Grafana) | Low | Mitigated | Done BL-021 local Grafana/Prometheus |

---

## Security риски

| ID | Риск | Вероятность | Impact | Митигация |
|----|------|-------------|--------|-----------|
| SR-01 | Admin API key in env — single point of failure | Medium | Critical | Rotate key; Admin UI exists; future RBAC |
| SR-02 | JWT secret compromise | Low | Critical | 32+ byte random, env-only |
| SR-03 | connection_config in DB — relay secrets | Medium | High | Encrypt at rest (future), access control |
| SR-04 | Android debug signing in release | Low | Mitigated | Done BL-013 — key.properties / production keystore path |
| SR-05 | No WAF/DDoS protection | Medium | Medium | Caddy rate limit, future Cloudflare |
| SR-06 | Telemetry user_id linkable | Low | Medium | No PII by design, retention policy TODO |

---

## Risk Matrix Summary

```
Impact ↑
Critical │ BR-01  IR-01  SR-01  TR-03
High     │ TR-02  BR-03  SR-03
Medium   │ BR-02  BR-04  IR-02  IR-05  SR-05
Low      │ SR-02  SR-06  (mitigated: TR-01,04,05,06 IR-03,04,06 SR-04)
         └────────────────────────────→ Probability
           Low    Medium    High
```

---

## Top 5 Actions

1. **Live-test ЮKassa** (BR-01) — only on explicit request; unblocks BL-030
2. **Device recheck** APK v0.1.1+17 on physical Android
3. **Off-site backup copy** (IR-03 residual) — optional hardening
4. **Real domain** (IR-02) before public launch
5. **Measured perf** T1–T4 (startup/connect/recovery/uptime) — docs/29, docs/30
