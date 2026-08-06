# StreamPass — Текущее состояние

> Дата: 2026-08-06 | Источник: код + `docs/04_Backlog.md` + prod

---

## Что работает

### Backend (Go 1.22)

| Модуль | Описание |
|--------|----------|
| **Auth** | Register/Login/Logout/Refresh, Argon2id, JWT, Redis sessions |
| **Rules / Config** | Версионированные правила и клиентский config (в т.ч. auto-update fields) |
| **Relay** | Реестр, health, ranking, region catalog `de/nl/pl/fi`, `?region=` |
| **Telemetry / Billing / Exclusions / Admin users** | Реализованы (billing без live ЮKassa) |
| **Health Monitor** | TCP-probe → `POST /servers/health` |
| **Metrics** | `/metrics` (Prometheus text; публично закрыт Caddy) |

### Docker Compose

`postgres`, `redis`, `backend`, `healthmonitor`, `caddy` (+ volume `./admin`), `prometheus`, `node-exporter`, `grafana`.

### Android Client (Flutter) — v0.1.1+25

| Возможность | Статус |
|-------------|--------|
| Ускоритель (Hysteria2 + TUN + protect) | ✅ |
| Decision Engine + Rule polling (default DIRECT) | ✅ |
| OS split-tunnel RU CIDR / intl routes | ✅ |
| App bypass (Госуслуги/ФНС/банки, эвристика) | ✅ |
| Split DNS (.ru → Yandex, foreign → DoH) | ✅ |
| Exclusions sync | ✅ |
| Refresh token | ✅ |
| Region / relay picker | ✅ |
| Soft/hard update prompt | ✅ |
| APK release signing (key.properties) | ✅ когда JKS на месте |

Диагностика DIRECT vs VPN: `docs/33_DirectVsVpnBypass.md`.

### Админка / ops

| Компонент | URL / путь | Статус |
|-----------|------------|--------|
| Admin UI | `https://…/admin/` | ✅ |
| Regions API | `GET /api/v1/regions` | ✅ |
| Backups | `/var/backups/streampass`, cron 03:00 UTC | ✅ |
| Grafana / Prometheus | `127.0.0.1:3000` / `:9090` | ✅ local-only |

### Тесты

- Go unit + `backend/internal/integrationtest/`  
- Flutter unit + `client/test/e2e_flow_test.dart` (mock API)  
- `scripts/SmokeTest.ps1`, `scripts/loadtest`  
- CI: `.github/workflows/ci.yml`

---

## Частично / остатки

| Компонент | Статус |
|-----------|--------|
| **Billing / ЮKassa** | Код есть; live-тест Skipped (BL-004); auto-renewal Blocked (BL-030) |
| **Physical device E2E** | Код/APK готовы; ручная проверка на устройстве |
| TCP underlay fallback | ✅ Done (BL-017): UDP 443→8443→24443, TCP underlay 8443/24443 |
| Off-site backup | ✅ Done (BL-035): encrypt → `157.167`; verify via `VerifyOffsiteBackup.ps1` |
| **nip.io** | Работает; брендовый домен нет |

---

## Не в scope сейчас

| Компонент | Статус |
|-----------|--------|
| Windows / iOS / macOS клиенты | Open (BL-023…025) — не начинать без запроса |
| Kubernetes / ML / multi-hop и пр. | Исключения ТЗ §21 |

---

## Prod

| | |
|--|--|
| API | `https://212-43-156-33.nip.io` |
| Admin | `https://212-43-156-33.nip.io/admin/` |
| Relays | `nl-native-1`, `nl-amsterdam-1` (регион `nl`); DE/PL/FI — софт готов, нод нет |
| APK | `StreamPass-v0.1.1+25-signed-arm64.apk` |
| OTA | `https://212-43-156-33.nip.io/downloads/StreamPass.apk` |

---

## Следующие шаги (приоритет)

1. Ручной device E2E на телефоне (adb)
2. Брендовый домен — только после покупки DNS
3. ЮKassa live — только по явному запросу → затем BL-030
4. Desktop/mobile другие ОС — только по явному запросу
