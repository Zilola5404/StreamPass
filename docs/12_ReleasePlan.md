# StreamPass — План релизов

> Дата: 2026-08-03

---

## MVP (текущий этап)

**Цель:** Первые пользователи могут зарегистрироваться, оплатить, подключиться одной кнопкой.

| Критерий | Статус |
|----------|--------|
| Backend API на VPS | ✅ Docker Compose |
| PostgreSQL + Redis | ✅ |
| Auth (register/login/logout) | ✅ |
| Rules + Config API | ✅ |
| Relay Manager + Health Monitor | ✅ |
| Billing (ЮKassa client) | ⚠️ Code ready, not live-tested |
| Android app UI | ✅ |
| VPN tunnel (Hysteria2) | ❌ Stub |
| Decision/Rule Engine (client) | ❌ |
| HTTPS | ✅ Caddy |
| CI/CD | ❌ |

**Target date:** TODO: Требуется уточнение

---

## Beta

**Цель:** Closed beta с 10-50 пользователями.

| Задача | Зависимость |
|--------|-------------|
| Working VPN tunnel end-to-end | MVP blocker |
| Live ЮKassa testing | MVP |
| Decision Engine on client | MVP |
| CI/CD pipeline | — |
| Integration tests | CI/CD |
| Production Android signing | — |
| Error monitoring (Sentry/similar) | — |
| Basic load testing | — |

**Target date:** TODO: Требуется уточнение

---

## Production v1.0

**Цель:** Public release, app store listing.

| Задача | Зависимость |
|--------|-------------|
| All MVP + Beta criteria met | Beta |
| Real domain (api.streampass.com) | — |
| Prometheus + Grafana | — |
| Backup automation | — |
| Security audit | — |
| Performance benchmarks (ТЗ §22) | — |
| Privacy policy + ToS | — |
| App Store / Google Play submission | Production signing |

**Performance targets (ТЗ §22):**
- Client startup ≤ 2s
- Connection ≤ 5s
- Auto-recovery ≤ 10s
- Server availability ≥ 99.9%

---

## Future (post-v1.0)

| Version | Features (из ТЗ §24) |
|---------|---------------------|
| v1.1 | Linux client, additional relays, improved Decision Engine |
| v1.2 | TUIC transport, advanced routing rules |
| v1.5 | Score Engine based on telemetry (no ML) |
| v2.0 | Custom transport, Multipath QUIC, adaptive routing |

**Not in scope (ТЗ §21):** Kubernetes, ML, MASQUE, ASN/GeoIP routing, Browser Extension, Corporate version, Multi-Hop.

---

## Release Checklist

- [ ] All tests green
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Docker images tagged
- [ ] Migration tested
- [ ] Rollback plan documented
- [ ] Smoke test on staging
