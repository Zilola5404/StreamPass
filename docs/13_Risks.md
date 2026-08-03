# StreamPass — Риски

> Дата: 2026-08-03

---

## Технические риски

| ID | Риск | Вероятность | Impact | Митигация |
|----|------|-------------|--------|-----------|
| TR-01 | VPN tunnel (go_core) не реализован — MVP заблокирован | High | Critical | BL-001, BL-002 — приоритет P0 |
| TR-02 | Hysteria2 dependency — breaking changes в upstream | Medium | High | Pin version, monitor releases |
| TR-03 | Custom JWT/Redis — bugs in security-critical code | Medium | Critical | Security review, add tests |
| TR-04 | No go.sum — non-reproducible builds | Medium | Medium | BL-027 |
| TR-05 | No integration tests — regressions undetected | High | High | BL-011, CI/CD |
| TR-06 | vendor-src/mobile missing — gomobile build fails | High | High | Vendor or install x/mobile |

---

## Бизнес риски

| ID | Риск | Вероятность | Impact | Митигация |
|----|------|-------------|--------|-----------|
| BR-01 | ЮKassa не протестирована — payments fail in production | High | Critical | Sandbox testing before launch |
| BR-02 | Subscription cancel semantics unclear | Medium | Medium | Product decision (ADR-010) |
| BR-03 | Обещание «ускорение интернета» — user disappointment | Medium | High | Positioning: «стабильный маршрут» (ТЗ CTO note) |
| BR-04 | Только Android — limited market | Medium | Medium | Roadmap: iOS, Windows |

---

## Инфраструктурные риски

| ID | Риск | Вероятность | Impact | Митигация |
|----|------|-------------|--------|-----------|
| IR-01 | Single VPS — no HA | High | Critical | Multi-relay, future k8s (post-MVP) |
| IR-02 | nip.io domain — not production-grade | High | Medium | Real domain before production |
| IR-03 | No automated backup | High | Critical | BL-033, docs/27 |
| IR-04 | No CI/CD — manual deploy errors | High | Medium | BL-010 |
| IR-05 | Redis no persistence (no save/appendonly) | Medium | Medium | Sessions lost on restart — acceptable for MVP |
| IR-06 | No monitoring (Prometheus/Grafana) | High | Medium | BL-021 |

---

## Security риски

| ID | Риск | Вероятность | Impact | Митигация |
|----|------|-------------|--------|-----------|
| SR-01 | Admin API key in env — single point of failure | Medium | Critical | Rotate key, future Admin Panel with RBAC |
| SR-02 | JWT secret compromise | Low | Critical | 32+ byte random, env-only |
| SR-03 | connection_config in DB — relay secrets | Medium | High | Encrypt at rest (future), access control |
| SR-04 | Android debug signing in release | High | High | BL-013 production keystore |
| SR-05 | No WAF/DDoS protection | Medium | Medium | Caddy rate limit, future Cloudflare |
| SR-06 | Telemetry user_id linkable | Low | Medium | No PII by design, retention policy TODO |

---

## Risk Matrix Summary

```
Impact ↑
Critical │ TR-01  BR-01  IR-01  SR-01
High     │ TR-02  TR-03  TR-05  BR-03  IR-03  SR-04
Medium   │ TR-04  TR-06  BR-02  BR-04  IR-02  IR-04  IR-05  IR-06  SR-03  SR-05
Low      │ SR-02  SR-06
         └────────────────────────────→ Probability
           Low    Medium    High
```

---

## Top 5 Actions

1. **Реализовать VPN tunnel** (TR-01) — unblock MVP
2. **Live-test ЮKassa** (BR-01) — unblock monetization
3. **CI/CD + integration tests** (TR-05, IR-04)
4. **Production signing + real domain** (SR-04, IR-02)
5. **Backup automation** (IR-03)
