# StreamPass

Клиентский VPN с умной маршрутизацией (DIRECT / RELAY) и общий Go-backend.

| | |
|---|---|
| **API (prod)** | `https://212-43-156-33.nip.io/api/v1` |
| **Клиент** | Flutter Android (`client/`), go_core → `streampasscore.aar` |
| **Backend** | Go 1.22 Clean Architecture (`backend/`) |
| **Статус** | Android MVP (код) готов — см. `docs/04_Backlog.md` / `docs/03_CurrentState.md` |

## Репозиторий

```
backend/          — API-сервер (auth, billing, relays, rules, exclusions…)
client/           — Flutter Android + go_core (Hysteria2 / Decision / DNS)
admin/            — Operator UI (/admin/)
shared/           — общий Go-код (+ region catalog)
infrastructure/   — Prometheus / Grafana provisioning
docs/             — ТЗ, backlog, API, architecture
ai/               — контекст для агентов (CurrentTask.md)
.github/workflows — CI: go test, flutter analyze/test, docker compose build
scripts/          — Backup, SmokeTest, LoadTest, …
```

## Backend — быстрый старт

```bash
cp .env.example .env   # DB_PASSWORD, JWT_SECRET, ADMIN_API_KEY
docker compose up -d --build
curl -s http://localhost:8080/health
```

Локально без Docker: Go 1.22+, PostgreSQL, Redis — см. `backend/config.example.yaml`.

```bash
go test ./...
go build -o streampass-server ./backend/cmd/server
```

## Android-клиент

```bash
cd client
flutter pub get
# после изменений в go_core — пересобрать AAR (нужен JDK / Android Studio JBR):
cd go_core && gomobile bind -target=android -androidapi=21 -o streampasscore.aar ./mobile
cp streampasscore.aar ../android/app/libs/
cd ..
flutter build apk --release --target-platform android-arm64
# APK: build/app/outputs/flutter-apk/app-release.apk (~18 MB)
```

Подробности go_core: `client/go_core/README.md`.

### Что умеет клиент сейчас

- Connect / disconnect VPN (Hysteria2 + TUN + `protect`)
- Decision Engine + hot-reload правил (`GET /rules`)
- DNS Cache + DoH
- Пользовательские исключения (`GET/PUT /exclusions`)
- UDP fallback портов 443 → 8443 → 24443
- Авто-refresh access token (`POST /refresh`)
- Выбор региона / relay (BL-026)
- Soft/hard update prompt по `GET /config` (BL-022)
- APK: `v0.1.1+17`

## Документация

| Файл | Содержание |
|------|------------|
| `docs/04_Backlog.md` | Задачи и статусы |
| `docs/08_API.md` | HTTP API |
| `docs/07_Architecture.md` | Архитектура |
| `docs/18_KnownLimitations.md` | Известные ограничения |
| `ai/CurrentTask.md` | Текущая задача агента |
| `docs/RelayServers.md` | Relay / Hysteria на VPS |

## CI

`.github/workflows/ci.yml` на `main` / PR:

1. `go test ./...` + build server  
2. `go test` в `client/go_core`  
3. `flutter analyze` + `flutter test`  
4. `docker compose build`

## Что ещё открыто

- Live-тест ЮKassa (BL-004 — Skipped) → auto-renewal (BL-030 Blocked)  
- Windows / iOS / macOS клиенты (BL-023…025)  
- Физический device E2E re-check APK +24  
- TCP underlay fallback (UDP ports уже есть, BL-017)  
- Брендовый домен вместо nip.io  

**Уже Done:** Admin UI, Prometheus/Grafana, DNS/DoH, release signing path, regions, backups, CI, Flutter E2E, loadtest, **app OS-bypass UI**, **OTA APK** (`/downloads/StreamPass.apk`).

Полный список: `docs/04_Backlog.md`.
