# Ответ на AUDIT-001 — StreamPass (фактическая кодбаза)

**Дата:** 4 августа 2026  
**Проверил:** по репозиторию `main` @ `2c56972` (Go backend + Flutter client)  
**Исходный аудит:** `reports/Audit/AUDIT-001-audit.md`

---

## Главный вывод

**AUDIT-001 не применим к текущему репозиторию StreamPass.**

Аудит описывает стек, которого **нет в проекте**:

| В аудите | В репозитории |
|----------|----------------|
| `server.ts` (Node.js Express) | **отсутствует** |
| React `src/` (Diagnostics.tsx, RouteTester.tsx) | **отсутствует** |
| `package.json`, `@google/genai`, `dotenv` | **отсутствует** |
| In-memory массивы `users`, `relays`, `activeSessions` | **PostgreSQL + Redis** |
| Пути `/api/...` без версии | **`/api/v1/...`** (`router.go`) |

Вероятная причина: аудит выполнен по **другому прототипу/демо** (веб-панель), а не по production-кодовой базе Go + Flutter.

**Пересмотренный вердикт по фактическому StreamPass:** замечания P0 из AUDIT-001 **не подтверждаются**. Критические блокеры релиза из аудита **не актуальны**. Реальные пробелы MVP зафиксированы в `docs/04_Backlog.md` (BL-005, BL-006 и др.).

---

## Разбор замечаний по пунктам

### §1 Несоответствия ТЗ

| # | Замечание аудита | Вердict | Обоснование |
|---|------------------|---------|-------------|
| 1.1 | CIDR не проверяются в `POST /api/rules/test-route` | **Не применимо** | Эндпоинта `test-route` нет в Go API и в ТЗ (`docs/02_TZ.md`). CIDR как тип правила **есть** в домене (`backend/internal/domain/rule/rule.go`, `KindCIDR`), валидация при публикации — в `rule/service.go`. **Decision Engine на клиенте не реализован** — это BL-005 (известное ограничение, `docs/18_KnownLimitations.md`). |
| 1.2 | TUN/Hysteria2 — только симуляция в Express/React | **Неверно** | Реальный TUN: `StreamPassVpnService.kt`, Go core `tunnel.go`, Hysteria2 client, AAR `streampasscore.aar`. |
| 1.3 | Нет cron для истечения подписки | **Частично неверно** | Статус вычисляется при каждом запросе: `subscription.NewInfo(activeUntil, now)` — если `active_until` в прошлом → `INACTIVE`. Отдельный cron для «разжалования тарифа» не обязателен для корректности API. Фоновая задача — опциональна (аналитика, уведомления), не P0. |

### §2 Архитектура

| # | Замечание | Верdict | Обоснование |
|---|-----------|---------|-------------|
| 2.1 | In-memory state | **Неверно** | Postgres: users, relays, rules, payments (`backend/internal/infrastructure/postgres/`). Redis: refresh-сессии с TTL (`session_store.go`). |
| 2.2 | Неиспользуемые npm-зависимости | **Не применимо** | Нет `package.json`. |
| 2.3 | Монолитный `server.ts` 620 строк | **Неверно** | Слои: domain / application / infrastructure / handler (`backend/internal/...`). |

### §3 Безопасность

| # | Замечание | Верdict | Обоснование |
|---|-----------|---------|-------------|
| 3.1 | Пароли в открытом виде | **Неверно** | Argon2id: `backend/internal/infrastructure/security/argon2_hasher.go`. Login: `hasher.Verify()` в `auth/login.go`. |
| 3.2 | Фиктивные JWT `jwt_token_${id}_${Date.now()}` | **Неверно** | HS256 JWS: `jwt_minimal.go`, секрет из `jwt.secret` / `JWT_SECRET` (`config.example.yaml`). Refresh в Redis с TTL. |
| 3.3 | Hardcoded `admin-secret-key` / `demo-admin` | **Неверно** | `middleware/admin.go` — ключ из конфига `admin.api_key` / `ADMIN_API_KEY`, сравнение через `subtle.ConstantTimeCompare`. |
| 3.4 | Нет валидации входных данных | **Частично верно** | Есть `auth/validation.go`, `validateRules()` для rule set. Можно усилить (Zod-подобные схемы на все POST) — **P2**, не дубликат P0 из аудита. |

### §4 Производительность

| # | Замечание | Верdict | Обоснование |
|---|-----------|---------|-------------|
| 4.1 | Утечка `activeSessions` без очистки | **Неверно** | Сессии в Redis с TTL = срок refresh-токена. |
| 4.2 | O(N) в test-route | **Не применимо** | Нет test-route; Decision Engine (BL-005) ещё не реализован. |
| 4.3 | Нет `useMemo` в React | **Не применимо** | UI — Flutter/Dart. |

### §5 Качество кода

| # | Замечание | Верdict | Обоснование |
|---|-----------|---------|-------------|
| 5.1 | Demo user на `/api/auth/me` без токена | **Не применимо** | Нет такого эндпоинта; auth через Bearer + `RequireAuth`. |
| 5.2 | Нет типизации Express | **Не применимо** | Go + Dart. |

### §6 Масштабируемость

| # | Замечание | Верdict | Обоснование |
|---|-----------|---------|-------------|
| 6.1 | Stateless невозможен | **Неверно** | Backend stateless; state в Postgres/Redis. |
| 6.2 | Нет `/api/v1` | **Неверно** | `apiV1Prefix = "/api/v1"` в `router.go`. |

### §7 Документация

| # | Замечание | Верdict | Обоснование |
|---|-----------|---------|-------------|
| 7.1 | Нет предупреждения о demo-токенах | **Устарело** | Актуальные `docs/` описывают Go backend, Argon2, env-секреты. |

---

## Что действительно остаётся открытым (не из AUDIT-001)

Эти пункты **уже в backlog**, не дублируют P0 аудита:

| ID | Задача | Статус |
|----|--------|--------|
| BL-005 | Decision Engine на клиенте (DIRECT/RELAY/FALLBACK, в т.ч. CIDR) | Open |
| BL-006 | Rule Engine — загрузка и применение правил | Open |
| BL-004 | Live-тест ЮKassa | Open |
| BL-011 | Integration tests backend (testcontainers) | Open |

---

## Рекомендации

1. **Пометить AUDIT-001 как INVALID / WRONG TARGET** или архивировать; не использовать для gate релиза.
2. **Провести повторный аудит** по фактическим путям:
   - `backend/` (Go)
   - `client/` (Flutter + `go_core/`)
   - `docs/02_TZ.md`, `docs/18_KnownLimitations.md`
3. **Не внедрять правки из AUDIT-001** (создание `server.ts`, bcrypt в Node и т.д.) — это создаст второй параллельный backend.

---

## Исправления в коде по этому review

**Не требуются** — замечания P0/P1 ссылаются на несуществующие файлы. Реальные улучшения — через BL-005/BL-006 и усиление валидации API (отдельная задача P2).
