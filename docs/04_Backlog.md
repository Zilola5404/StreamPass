# StreamPass — Backlog

> Дата: 2026-08-08 | Формат: ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный

---

## P0 — Критично для MVP

| ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный |
|----|----------|----------|-----------|--------|-------------|---------------|
| BL-001 | Hysteria2 tunnel в go_core | Техническая перепроверка и hardening существующего Hysteria2 transport в `client/go_core/mobile/tunnel.go`: реальный relay handshake/data path, TUN lifecycle, protected underlay, fallback и Android E2E evidence; использовать готовую Hysteria2 реализацию | P0 | In Progress | — | Developer + QA |
| BL-002 | Подключить streampasscore.aar | Собрать и положить в Android libs, проверить TunnelBridge | P0 | Done | BL-001 | 2026-08-03 |
| BL-003 | End-to-end VPN на Android | Connect → TUN → Hysteria2 → relay, проверка IP | P0 | Done* | BL-001, BL-002 | 2026-08-03 |
| BL-004 | Live-тест ЮKassa | Sandbox ключи, CreatePayment + webhook flow | P0 | Skipped | — | 2026-08-04 |
| BL-005 | Decision Engine (клиент) | DIRECT/RELAY/FALLBACK по правилам | P0 | Done | BL-003 | 2026-08-04 |
| BL-006 | Rule Engine (клиент) | Загрузка правил с GET /api/v1/rules, polling, hot-reload | P0 | Done | BL-005 | 2026-08-04 |

> **BL-001 audit note (2026-08-08):** базовый transport уже реализован и ранее отмечен Done. Team Lead возвращает задачу в `In Progress`, потому что для окончательной приёмки требуется воспроизводимое evidence полного физического Android E2E и проверки реального TCP/UDP data path, lifecycle и security gates. BL-003 исторически закрыт, но его evidence используется как часть текущего gate; `Done*` означает «исторически закрыт, evidence подлежит повторной проверке в рамках BL-001».

## P1 — Важно

| ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный |
|----|----------|----------|-----------|--------|-------------|---------------|
| BL-010 | CI/CD GitHub Actions | build, test, docker compose build | P1 | Done | — | 2026-08-04 |
| BL-011 | Integration tests backend | auth, billing, relay с Postgres testcontainers | P1 | Done | BL-010 | 2026-08-04 |
| BL-012 | Обновить README | Исправить статус Health Monitor, добавить client docs | P1 | Done | — | 2026-08-04 |
| BL-013 | Release signing Android | Production keystore вместо debug | P1 | Done | BL-003 | 2026-08-04 |
| BL-014 | Exclusions sync | Синхронизация пользовательских исключений с backend | P1 | Done | BL-006 | 2026-08-04 |
| BL-015 | Refresh token rotation | Клиент: авто-обновление access token | P1 | Done | — | 2026-08-04 |
| BL-016 | DNS Cache + DoH | Локальный DNS на клиенте (ТЗ §7) | P1 | Done | BL-005 | 2026-08-04 |
| BL-017 | Fallback Strategy | UDP 443 → 8443 → 24443, затем TCP underlay 8443 → 24443 | P1 | Done | BL-001 | 2026-08-06 |

## P2 — Желательно

| ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный |
|----|----------|----------|-----------|--------|-------------|---------------|
| BL-020 | Admin Panel UI | Web UI для операторов вместо X-Admin-Key | P2 | Done | — | 2026-08-04 |
| BL-021 | Prometheus + Grafana | Мониторинг (ТЗ §18) | P2 | Done | — | 2026-08-04 |
| BL-022 | Auto Update клиента | OTA обновления APK | P2 | Done | BL-003 | 2026-08-04 |
| BL-023 | Windows клиент | WFP adapter (ТЗ §3) | P2 | Open | BL-001 | TODO |
| BL-024 | iOS клиент | Network Extension (ТЗ §3) | P2 | Open | BL-001 | TODO |
| BL-025 | macOS клиент | Network Extension (ТЗ §3) | P2 | Open | BL-001 | TODO |
| BL-026 | Несколько relay регионов | Frankfurt, Amsterdam, Warsaw, Helsinki | P2 | Done | — | 2026-08-04 |
| BL-027 | go.sum в репозитории | Добавить go.sum для reproducible builds | P2 | Done | — | 2026-08-04 |

## P3 — Backlog

| ID | Название | Описание | Приоритет | Статус | Зависимости | Ответственный |
|----|----------|----------|-----------|--------|-------------|---------------|
| BL-030 | Subscription auto-renewal | Автопродление подписки через ЮKassa | P3 | Blocked | BL-004 (Skipped) | — |
| BL-031 | E2E tests Flutter | Integration tests с mock backend | P3 | Done | BL-003 | 2026-08-04 |
| BL-032 | Load tests API | k6/vegeta на /api/v1 | P3 | Done | BL-010 | 2026-08-04 |
| BL-033 | Backup automation | Postgres backup cron | P3 | Done | — | 2026-08-04 |
| BL-034 | App OS-bypass UI | Выбор приложений без VPN + OTA APK URL | P2 | Done | BL-005 | 2026-08-06 |
| BL-035 | Off-site backup | Шифрованная копия дампов на второй хост + pull на PC | P3 | Done | BL-033 | 2026-08-06 |

## P1 — SaaS / Functional Spec gaps (`docs/02.2_FunctionalSpecification.md`)

> Источник: FS v1.0 Must/Should. Чеклист с привязкой к экранам: `docs/35_FS_ImplementationChecklist.md`.

| ID | Название | Описание (из FS) | Приоритет | Статус | Зависимости | Ответственный |
|----|----------|------------------|-----------|--------|-------------|---------------|
| BL-040 | Live ЮKassa + возврат в app | Ранее BL-004: live/sandbox оплата, poll статуса на E06, deep-link return | P0 | Blocked | — | resume poll Done; **live keys отсутствуют** |
| BL-041 | Logout + session UX | E05/E10: «Выйти», clear tokens, Stop tunnel → E01; единые тексты сессии | P1 | Done | — | 2026-08-06 |
| BL-042 | Сброс пароля | E01 «Забыли пароль?» + backend flow | P1 | Done | — | 2026-08-06 |
| BL-043 | Профиль и удаление аккаунта | E10: просмотр email, смена пароля, удалить аккаунт (двойное подтверждение) | P1 | Done | BL-041 | 2026-08-06 |
| BL-044 | Статистика (реальные метрики) | E03: online time, avg RTT, reconnects; без URL | P1 | Done | — | 2026-08-06 |
| BL-045 | Синхрон Auto Mode ↔ Автовыбор | E02 переключатель = `autoSelectRelay` (убрать UI-only флаг) | P1 | Done | — | 2026-08-06 |
| BL-046 | Reconnect при смене сервера | E04→E02: если Connected — Connecting к новому relay | P1 | Done | BL-017 | 2026-08-06 |
| BL-047 | UX автосмены relay | При деградации — смена + toast «Переключили сервер…» | P1 | Done | BL-017 | 2026-08-06 |
| BL-048 | Тарифы и история платежей | E06: месяц/год, история; политика доступа до `active_until` после отмены | P1 | Done | BL-040 | 2026-08-06 (API+UI; live pay ждёт BL-040) |
| BL-049 | Устройства / лимит | E10 список устройств, revoke, лимит (**Should→P2**) | P2 | Open | BL-043 | TODO |
| BL-050 | Admin: Premium / бан / audit | Users: выдать/забрать подписку, поиск, audit log | P2 | Open | BL-020 | TODO |
| BL-051 | Уведомления о сбоях | E05 toggle + системные уведомления при обрыве | P2 | Open | — | TODO |
| BL-052 | Язык / тема / О приложении | E05 Should-секции | P3 | Open | — | TODO |
| BL-053 | Device SLA measurement | Замер cold start ≤2с, connect ≤5с, recover ≤10с (ТЗ §22) | P1 | Done | — | 2026-08-06 script+unit; device connect/recover manual |
| BL-054 | Terms / Privacy на E01 | Ссылки при регистрации | P2 | Open | — | TODO |

---

## Легенда статусов

- **Open** — не начато
- **In Progress** — в работе
- **Done** — завершено и подтверждено
- **Blocked** — заблокировано внешней зависимостью
- **Skipped** — сознательно не выполняется
- `Done*` — исторически закрыто, но evidence используется для текущей перепроверки

## Легенда приоритетов

- **P0** — блокер MVP
- **P1** — важно для релиза
- **P2** — post-MVP
- **P3** — backlog
