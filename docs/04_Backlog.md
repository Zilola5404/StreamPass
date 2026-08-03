# StreamPass — Backlog

> Дата: 2026-08-03 | Формат: ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный

---

## P0 — Критично для MVP

| ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный |
|----|----------|----------|-----------|--------|-------------|---------------|
| BL-001 | Hysteria2 tunnel в go_core | Реализовать transport в `client/go_core/mobile/tunnel.go`, собрать AAR через gomobile | P0 | Done | — | 2026-08-03 |
| BL-002 | Подключить streampasscore.aar | Собрать и положить в Android libs, проверить TunnelBridge | P0 | Done | BL-001 | 2026-08-03 |
| BL-003 | End-to-end VPN на Android | Connect → TUN → Hysteria2 → relay, проверка IP | P0 | Done | BL-001, BL-002 | 2026-08-03 |
| BL-004 | Live-тест ЮKassa | Sandbox ключи, CreatePayment + webhook flow | P0 | Open | — | TODO |
| BL-005 | Decision Engine (клиент) | DIRECT/RELAY/FALLBACK по правилам | P0 | Done | BL-003 | 2026-08-04 |
| BL-006 | Rule Engine (клиент) | Загрузка правил с GET /api/v1/rules, polling, hot-reload | P0 | Done | BL-005 | 2026-08-04 |

## P1 — Важно

| ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный |
|----|----------|----------|-----------|--------|-------------|---------------|
| BL-010 | CI/CD GitHub Actions | build, test, docker push | P1 | Open | — | TODO |
| BL-011 | Integration tests backend | auth, billing, relay с Postgres testcontainers | P1 | Open | BL-010 | TODO |
| BL-012 | Обновить README | Исправить статус Health Monitor, добавить client docs | P1 | Open | — | TODO |
| BL-013 | Release signing Android | Production keystore вместо debug | P1 | Open | BL-003 | TODO |
| BL-014 | Exclusions sync | Синхронизация пользовательских исключений с backend | P1 | Open | BL-006 | TODO |
| BL-015 | Refresh token rotation | Клиент: авто-обновление access token | P1 | Open | — | TODO |
| BL-016 | DNS Cache + DoH | Локальный DNS на клиенте (ТЗ §7) | P1 | Open | BL-005 | TODO |
| BL-017 | Fallback Strategy | UDP 443 → 8443 → 24443 → TCP (ТЗ §10) | P1 | Open | BL-001 | TODO |

## P2 — Желательно

| ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный |
|----|----------|----------|-----------|--------|-------------|---------------|
| BL-020 | Admin Panel UI | Web UI для операторов вместо X-Admin-Key | P2 | Open | — | TODO |
| BL-021 | Prometheus + Grafana | Мониторинг (ТЗ §18) | P2 | Open | — | TODO |
| BL-022 | Auto Update клиента | OTA обновления APK | P2 | Open | BL-003 | TODO |
| BL-023 | Windows клиент | WFP adapter (ТЗ §3) | P2 | Open | BL-001 | TODO |
| BL-024 | iOS клиент | Network Extension (ТЗ §3) | P2 | Open | BL-001 | TODO |
| BL-025 | macOS клиент | Network Extension (ТЗ §3) | P2 | Open | BL-001 | TODO |
| BL-026 | Несколько relay регионов | Frankfurt, Amsterdam, Warsaw, Helsinki | P2 | Open | — | TODO |
| BL-027 | go.sum в репозитории | Добавить go.sum для reproducible builds | P2 | Open | — | TODO |

## P3 — Backlog

| ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный |
|----|----------|----------|-----------|--------|-------------|---------------|
| BL-030 | Subscription auto-renewal | Автопродление подписки через ЮKassa | P3 | Open | BL-004 | TODO |
| BL-031 | E2E tests Flutter | Integration tests с mock backend | P3 | Open | BL-003 | TODO |
| BL-032 | Load tests API | k6/vegeta на /api/v1 | P3 | Open | BL-010 | TODO |
| BL-033 | Backup automation | Postgres backup cron | P3 | Open | — | TODO |

---

## Легенда статусов

- **Open** — не начато
- **In Progress** — в работе
- **Done** — завершено
- **Blocked** — заблокировано

## Легенда приоритетов

- **P0** — блокер MVP
- **P1** — важно для релиза
- **P2** — post-MVP
- **P3** — backlog
