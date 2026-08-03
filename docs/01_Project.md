# StreamPass — Паспорт проекта

> Версия: MVP v0.1 | Дата: 2026-08-03

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
| Подписка и оплата (ЮKassa) | ⚠️ Backend готов, live-тест не проводился |
| Получение правил и конфигурации | ✅ Backend API |
| Список relay-серверов | ✅ Backend API + Android UI |
| VPN-подключение (Hysteria2) | ❌ Go core — stub |
| Decision Engine / Rule Engine на клиенте | ❌ Не реализовано |
| Телеметрия (без PII) | ✅ Backend + Android UI |
| Admin Panel UI | ❌ Только admin API key |
| Windows / macOS / iOS клиенты | ❌ Не реализовано |

## Платформа

| Компонент | Технология | Статус |
|-----------|------------|--------|
| Backend API | Go 1.22, Clean Architecture | ✅ Работает |
| Database | PostgreSQL 16 | ✅ Миграции автоматические |
| Cache/Sessions | Redis 7 | ✅ |
| Mobile Client | Flutter (Dart 3.3+), Android VPNService | ⚠️ UI готов, туннель — stub |
| Client Core | Go (gomobile) | ❌ Stub |
| Reverse Proxy | Caddy 2 | ✅ Docker Compose |
| Health Monitor | Go worker | ✅ Docker Compose |
| CI/CD | — | ❌ Не настроен |

## Текущая стадия

**MVP — Backend + Android UI, туннель не завершён**

- Backend API развёрнут и функционален (Docker Compose)
- Flutter Android-клиент с экранами onboarding, home, servers, subscription, settings
- VPN-туннель (Hysteria2 через go_core) — заглушка
- Decision/Rule Engine на клиенте — не реализован

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
- Android Kotlin VPNService + MethodChannel

### Infrastructure
- Docker Compose: postgres, redis, backend, healthmonitor, caddy
- Домен: `212-43-156-33.nip.io` (Caddy TLS)

## Автор / Команда

TODO: Требуется уточнение

## Дата начала

Первый commit: `9570b95` — «Первый релиз StreamPass MVP»

## Статус

| Метрика | Значение |
|---------|----------|
| Backend | ~80% MVP |
| Android UI | ~55% MVP |
| VPN Tunnel | ~5% (stub) |
| CI/CD | 0% |
| Admin Panel | 0% |

## Репозиторий

- Путь: `C:\01_Projects\StreamPass`
- Ветка: `main`
- Remote: `origin/main`
