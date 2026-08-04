# StreamPass — Паспорт проекта

> Версия: MVP v0.1.1+17 | Дата: 2026-08-05

## Название

**StreamPass** — система интеллектуальной маршрутизации интернет-трафика.

## Назначение

Автоматически выбирать наиболее надёжный маршрут для каждого соединения, повышая стабильность работы зарубежных сервисов при нестабильной мобильной сети.

## Проблема пользователя

- Зарубежные сервисы (YouTube, Google, GitHub, OpenAI) нестабильно работают через мобильную сеть
- Российские сервисы (банки, госуслуги) должны работать напрямую
- Ручная настройка VPN/прокси сложна и ненадёжна

## Решение

Одна кнопка «Подключить» → система сама:
- выбирает оптимальный relay-сервер
- применяет правила маршрутизации (DIRECT / RELAY)
- переключает сервер при деградации
- обновляет правила и конфигурацию с backend

## Целевая аудитория

Пользователи мобильных устройств (Android в MVP), которым нужен стабильный доступ к зарубежным сервисам без ручной настройки.

## Основные функции (по ТЗ)

| Функция | Статус |
|---------|--------|
| Регистрация / авторизация | ✅ Backend + Android UI |
| Подписка и оплата (ЮKassa) | ⚠️ Backend готов, live-тест Skipped (BL-004) |
| Получение правил и конфигурации | ✅ Backend API + client polling |
| Список relay-серверов / регионов | ✅ Backend API + Android region picker |
| VPN-подключение (Hysteria2) | ✅ Go core + AAR + Android VPNService |
| Decision Engine / Rule Engine на клиенте | ✅ go_core + hot-reload |
| Телеметрия (без PII) | ✅ Backend + Android UI |
| Admin Panel UI | ✅ `/admin/` |
| Windows / macOS / iOS клиенты | ❌ Open (BL-023…025) |

## Платформа

| Компонент | Технология | Статус |
|-----------|------------|--------|
| Backend API | Go 1.22, Clean Architecture | ✅ Работает |
| Database | PostgreSQL 16 | ✅ Миграции автоматические |
| Cache/Sessions | Redis 7 | ✅ |
| Mobile Client | Flutter (Dart 3.3+), Android VPNService | ✅ v0.1.1+17 |
| Client Core | Go (gomobile) | ✅ Hysteria2 + Decision Engine |
| Reverse Proxy | Caddy 2 | ✅ Docker Compose |
| Health Monitor | Go worker | ✅ Docker Compose |
| Admin UI | Static `/admin/` | ✅ |
| Monitoring | Prometheus + Grafana | ✅ local-only |
| CI/CD | GitHub Actions | ✅ `.github/workflows/ci.yml` |

## Текущая стадия

**MVP — Backend + Android VPN end-to-end (prod)**

- Backend API развёрнут: `https://212-43-156-33.nip.io`
- Flutter Android-клиент: connect, rules, regions, exclusions, auto-update
- VPN-туннель (Hysteria2 через go_core) — работает
- Multi-region software ready; в prod — NL nodes only
- Open intentional: Windows / iOS / macOS; ЮKassa live Skipped

## Используемые технологии

### Backend
- Go 1.22.2
- PostgreSQL 16 (lib/pq)
- Redis 7 (custom RESP2 client)
- JWT HS256 (custom), Argon2id
- Docker, Docker Compose, Caddy

### Client
- Flutter 3.x, Dart >=3.3.0
- Provider, http, shared_preferences, google_fonts
- Android Kotlin VPNService + MethodChannel + streampasscore.aar

### Infrastructure
- Docker Compose: postgres, redis, backend, healthmonitor, caddy, prometheus, grafana
- Домен: `212-43-156-33.nip.io` (Caddy TLS)
- Daily Postgres backups (cron)

## Автор / Команда

TODO: Требуется уточнение

## Дата начала

Первый commit: `9570b95` — «Первый релиз StreamPass MVP»

## Статус

| Метрика | Значение |
|---------|----------|
| Backend | ~95% MVP |
| Android UI + VPN | ~90% MVP |
| VPN Tunnel | Done (BL-001…003) |
| CI/CD | Done (BL-010) |
| Admin Panel | Done (BL-020) |
| Monitoring / Backup | Done (BL-021, BL-033) |

## Репозиторий

- Путь: `C:\01_Projects\StreamPass`
- Ветка: `main`
- Remote: `origin/main`
