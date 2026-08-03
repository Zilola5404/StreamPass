# StreamPass — AI Context

> **Самый важный файл для AI-агентов.** Прочитать первым после `00_ProjectRules.md`.  
> Дата: 2026-08-03

---

## Проект

**StreamPass** — система интеллектуальной маршрутизации интернет-трафика.  
MVP: Go backend + Flutter Android client. VPN через Hysteria2 (planned, stub).

**Repo:** `C:\01_Projects\StreamPass` | **Branch:** `main`

---

## Цель

Пользователь нажимает «Подключить» → система автоматически выбирает relay, применяет правила DIRECT/RELAY, обеспечивает стабильный доступ к зарубежным сервисам.

**Не обещать «ускорение интернета»** — только «стабильный маршрут».

---

## Архитектура (кратко)

```
Flutter Android UI
  → auth_service, streampass_api, vpn_channel
  → Android VPNService → TunnelBridge → go_core (STUB)
  → HTTPS → Caddy → Go Backend :8080
  → PostgreSQL 16 + Redis 7
  → Hysteria2 Relay servers (external VPS)
  → Health Monitor worker (TCP probe)
```

Clean Architecture backend: `domain` → `application` → `infrastructure` → `http`  
Composition root: `backend/cmd/server/main.go`

Подробнее: `docs/07_Architecture.md`

---

## Стек

| Layer | Tech |
|-------|------|
| Backend | Go 1.22.2, stdlib HTTP, lib/pq, custom JWT/Redis/YAML |
| DB | PostgreSQL 16, Redis 7 |
| Mobile | Flutter 3.x, Dart 3.3+, Kotlin VPNService |
| Tunnel | Go gomobile (stub), Hysteria2 (planned) |
| Infra | Docker Compose, Caddy 2, distroless |
| Payments | ЮKassa HTTP client (not live-tested) |

---

## Главные файлы

| Файл | Зачем |
|------|-------|
| `backend/cmd/server/main.go` | DI, startup, migrations |
| `backend/internal/infrastructure/http/router/router.go` | All API routes |
| `backend/internal/infrastructure/postgres/migrations/` | DB schema |
| `backend/config.example.yaml` | Backend config template |
| `.env.example` | Secrets template |
| `docker-compose.yml` | Full stack deploy |
| `client/lib/main.dart` | Flutter entry, API URL |
| `client/lib/services/streampass_api.dart` | REST client |
| `client/lib/services/auth_service.dart` | Token storage |
| `client/lib/services/vpn_channel.dart` | Native VPN bridge |
| `client/android/.../StreamPassVpnService.kt` | TUN interface |
| `client/go_core/mobile/tunnel.go` | Tunnel STUB |
| `shared/config/` | YAML loader |
| `shared/errors/` | Unified error codes |
| `README.md` | Backend quick start (note: Health Monitor section outdated) |

---

## Правила изменения кода

1. **Clean Architecture** — business logic только в `application/` и `domain/`
2. **DI через конструкторы** — wiring в `main.go`
3. **API versioning** — все routes через `/api/v1/`
4. **Unified errors** — `shared/errors` + `httpx.WriteError`
5. **No secrets in code** — только `${ENV}` в config
6. **Minimal deps** — prefer stdlib, document new deps in ADR
7. **Tests** — `go test ./...` must pass
8. **Update docs** после изменений API/архитектуры/БД

---

## Что нельзя делать

- ❌ Ломать существующие API endpoints без versioning
- ❌ Добавлять PII/URL в telemetry
- ❌ Хардкодить секреты
- ❌ Вводить микросервисы, Kubernetes, ML (MVP scope)
- ❌ Придумывать несуществующие компоненты в документации
- ❌ Коммитить без явного запроса пользователя
- ❌ Менять код при задачах «только документация»

---

## Текущая задача

**Инициализация AI-friendly документации** — завершена 2026-08-03.

См. `ai/CurrentTask.md` для актуальной задачи разработки.

---

## Следующая задача

**BL-001: Реализовать Hysteria2 tunnel в go_core** — P0 blocker для MVP.

См. `ai/NextTask.md`, `docs/04_Backlog.md`.

---

## Быстрый старт для AI

```bash
# 1. Прочитать
docs/00_ProjectRules.md    # правила
docs/14_AIContext.md       # этот файл
docs/03_CurrentState.md    # что работает / что нет
ai/CurrentTask.md          # текущая задача

# 2. Проверить сборку
go build ./...
go test ./...

# 3. Client
cd client && flutter analyze && flutter test

# 4. Docker (если нужен full stack)
cp .env.example .env  # заполнить секреты
docker compose up -d --build
curl http://localhost:8080/health
```

---

## Документы по приоритету

1. `docs/14_AIContext.md` — этот файл
2. `docs/03_CurrentState.md` — реальное состояние
3. `docs/08_API.md` — API reference
4. `docs/07_Architecture.md` — архитектура
5. `docs/04_Backlog.md` — задачи
6. `docs/02_TZ.md` — полная спецификация продукта
7. `docs/16_AI_HANDOFF.md` — передача между AI-сессиями
