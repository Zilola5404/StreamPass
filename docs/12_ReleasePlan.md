# StreamPass — План релизов

> Дата: 2026-08-05

---

## MVP (текущий этап — почти закрыт)

**Цель:** Пользователи могут зарегистрироваться, (опционально) оплатить, подключиться одной кнопкой.

| Критерий | Статус |
|----------|--------|
| Backend API на VPS | ✅ `https://212-43-156-33.nip.io` |
| PostgreSQL + Redis | ✅ |
| Auth (register/login/logout/refresh) | ✅ |
| Rules + Config API | ✅ |
| Relay Manager + Health Monitor | ✅ |
| Regions API + multi-region software | ✅ (prod: NL nodes only) |
| Billing (ЮKassa client) | ⚠️ Code ready; live Skipped (BL-004) |
| Android app UI | ✅ v0.1.1+25 |
| VPN tunnel (Hysteria2) | ✅ |
| Decision/Rule Engine (client) | ✅ |
| HTTPS | ✅ Caddy |
| Admin UI | ✅ `/admin/` |
| CI/CD | ✅ GitHub Actions |
| Monitoring | ✅ Grafana/Prometheus local |
| Daily backups | ✅ BL-033 |

**Target date:** TODO: Требуется уточнение (device recheck + optional ЮKassa)

---

## Beta

**Цель:** Closed beta с 10-50 пользователями.

| Задача | Статус / зависимость |
|--------|----------------------|
| Working VPN tunnel end-to-end | ✅ Done |
| Live ЮKassa testing | ⏭️ Skipped / on request |
| Decision Engine on client | ✅ Done |
| CI/CD pipeline | ✅ Done |
| Integration tests | ✅ Done |
| Production Android signing | ✅ Path Done (BL-013) |
| Monitoring (Prometheus/Grafana) | ✅ Done |
| Basic load testing | ✅ Done (BL-032) |
| Closed beta users | Open |

**Target date:** TODO: Требуется уточнение

---

## Production v1.0

**Цель:** Public release, app store listing.

| Задача | Зависимость |
|--------|-------------|
| All MVP + Beta criteria met | Beta users + ЮKassa if monetizing |
| Real domain (api.streampass.com) | — |
| Off-site backup copy | Local cron Done |
| Security audit | — |
| Performance benchmarks measured (ТЗ §22) | — |
| Privacy policy + ToS | — |
| App Store / Google Play submission | Production signing path ready |

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

**Platform Open (intentional):** Windows (BL-023), iOS (BL-024), macOS (BL-025).

---

## Release Checklist

- [x] CI tests green
- [x] Documentation updated
- [x] CHANGELOG updated
- [ ] Docker images tagged
- [ ] Migration tested on rollback
- [ ] Rollback plan documented
- [x] Smoke test on prod (`scripts/SmokeTest.ps1`)
