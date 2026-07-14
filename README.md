# StreamPass — серверная часть (backend)

Бэкенд реализован на Go 1.22, по Clean Architecture (domain / application /
infrastructure), с Dependency Injection через конструкторы. Все внешние
зависимости, где это было возможно, реализованы без сторонних библиотек
(dependency-free) — см. раздел «Технические решения» ниже.

## Быстрый запуск на сервере (Docker Compose)

Предполагается чистый Linux-сервер с установленным Docker и Docker Compose.

```bash
git clone <ваш-репозиторий> streampass
cd streampass

cp .env.example .env
# отредактируйте .env: DB_PASSWORD, JWT_SECRET, ADMIN_API_KEY обязательны;
# YOOKASSA_* — при подключении реальной оплаты

docker compose up -d --build
```

После запуска:
- API доступно на `http://<сервер>:8080`
- Миграции БД применяются автоматически при старте (`postgres.Migrate`)
- Проверка живости: `GET /health`

Остановить: `docker compose down` (данные Postgres сохраняются в volume
`postgres_data`; `docker compose down -v` — удалить и их).

## Запуск без Docker (локальная разработка)

Требуется Go 1.22+, PostgreSQL, Redis.

```bash
cd backend
cp config.example.yaml config.yaml
export DB_HOST=localhost DB_NAME=streampass DB_USER=streampass DB_PASSWORD=... \
       REDIS_ADDR=localhost:6379 JWT_SECRET=... ADMIN_API_KEY=...
go run ./cmd/server
```

## Тесты и проверка сборки

```bash
go build ./...
go vet ./...
go test ./...
```

Всё сейчас зелёное: сборка без ошибок, `go vet` без замечаний, юнит-тесты
проходят для всех пакетов, где они есть (см. «Что покрыто тестами» ниже).

## Что реализовано

Полностью, со слоями domain → application → infrastructure → HTTP:

| Модуль | Что есть |
|---|---|
| **Auth** | Регистрация, логин, логаут. Argon2id-хеширование паролей, JWT (access+refresh), сессии в Redis с revoke |
| **Rule Service** | Версионированные наборы правил (`GET /rules`, `POST /rules` — admin) |
| **Config Service** | Версионированная динамическая конфигурация клиента (`GET /config`, `POST /config` — admin) |
| **Relay Manager** | Реестр relay-серверов, выбор лучшего (`GET /servers`), приём health-check (`POST /servers/health` — admin/internal) |
| **Telemetry** | Приём технических метрик без PII (`POST /telemetry`, требует auth) |
| **Billing** | Создание платежа, webhook, статус/отмена подписки. Провайдер абстрагирован интерфейсом `PaymentProvider`; есть рабочий HTTP-клиент для ЮKassa (`internal/infrastructure/payment/yookassa`) |

Инфраструктура: PostgreSQL (миграции применяются автоматически при
старте), Redis (сессии), JWT-аутентификация, rate limiting по IP,
структурированное JSON-логирование, единый формат ошибок на всех
эндпоинтах.

## Что НЕ реализовано / требует доработки

- **Admin Panel** — отдельный UI/сервис для операторов не строился;
  вместо него admin-эндпоинты (`POST /rules`, `POST /config`,
  `POST /servers/health`) защищены статическим ключом `X-Admin-Key`
  (см. `ADMIN_API_KEY`). Для полноценной админки нужен отдельный проект.
- **ЮKassa** — клиент написан по документации API, но ни разу не
  вызывался против реального аккаунта ЮKassa (в песочнице разработки нет
  сетевого доступа к `api.yookassa.ru`). Перед продакшеном обязательно
  протестировать `CreatePayment`/`FetchPaymentStatus` на реальных
  тестовых ключах ЮKassa.
- **Health Monitor** как отдельный воркер, который сам опрашивает relay
  и дёргает `POST /servers/health`, не реализован — реализован только
  приёмный конец (Relay Manager хранит и отдаёт результаты). Сам
  внешний health-checker — отдельная маленькая программа/cron, которую
  стоит добавить следующим шагом.
- **CI/CD** — пайплайн не настраивался в этой сессии.
- Отмена подписки (`POST /subscription/cancel`) реализована как
  немедленное прекращение доступа (`ActiveUntil = now`), а не как
  «остановить будущее автосписание» — в спецификации не было деталей
  про автопродление, это стоит уточнить с продуктом перед продакшеном.

## Технические решения (для ревью)

Песочница разработки не имеет доступа к `proxy.golang.org`, поэтому там,
где сторонняя библиотека была нужна, а простого решения на stdlib не
было, зависимости подключены как **локальные `replace`-модули**,
склонированные с GitHub (`vendor-src/`):

- `golang.org/x/crypto` (только для `argon2`) — GitHub-зеркало
- `golang.org/x/sys` (только для `cpu`, транзитивная зависимость argon2)
- `github.com/lib/pq` — чистый Go-драйвер Postgres, без транзитивных зависимостей

Там, где зависимость была небольшой и хорошо очерченной, вместо
вендоринга написана dependency-free реализация:

- **YAML-конфиг** (`shared/config`) — минимальный парсер под то
  подмножество YAML, которое реально используется в конфигах проекта
- **JWT** (`backend/internal/infrastructure/security/jwt_minimal.go`) —
  HS256, только нужный набор claim'ов
- **Redis-клиент** (`backend/internal/infrastructure/redisclient`) —
  RESP2 поверх `net.Conn`, только нужные команды (SET/GET/DEL/EXISTS/SCAN)
- **HTTP-роутер** — стандартный `http.ServeMux` (в Go 1.22+ он умеет
  `"POST /rules"`-паттерны из коробки, отдельный роутер не нужен)

Обоснование в каждом случае — в комментариях соответствующих файлов.

## Что покрыто тестами

Юнит-тесты есть для: `shared/config`, Argon2/JWT (`infrastructure/security`),
`infrastructure/redisclient` (RESP-парсер против мок-сервера),
`infrastructure/http/middleware` (rate limiter), `application/rule`,
`application/configsvc`.

Не покрыты тестами (не хватило времени в рамках этой сессии, стоит
добавить в первую очередь): `application/auth`, `application/billing`,
`application/relay`, `application/telemetry`, Postgres-репозитории
(нужны интеграционные тесты с реальной БД, например через `testcontainers`
или `dockertest` — в песочнице разработки Docker недоступен, поэтому не
написаны здесь), HTTP-хендлеры (нужны end-to-end тесты роутера).

## Структура проекта

```
shared/                          — общий код (errors, logger, config, idgen)
backend/
  cmd/server/main.go             — composition root
  internal/
    domain/                      — доменные модели по bounded context'ам
    application/                 — use cases / сервисы
    infrastructure/
      postgres/                  — репозитории + миграции (go:embed)
      redisclient/                — Redis-клиент + SessionStore
      security/                   — Argon2, JWT
      payment/yookassa/           — клиент ЮKassa
      http/                       — router, middleware, handlers
  migrations/ → internal/infrastructure/postgres/migrations/
  config.example.yaml
  Dockerfile
docker-compose.yml
.env.example
```

## Дальнейшая работа

Учитывая объём и сложность оставшихся задач (Admin Panel, боевое
тестирование ЮKassa, интеграционные тесты, CI/CD), для продолжения
рекомендуется **Claude Code** — там можно итеративно гонять `go test`
и `docker compose up` с реальной обратной связью на каждом шаге, а не
переносить готовый код через чат.
