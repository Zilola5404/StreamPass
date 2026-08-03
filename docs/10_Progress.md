# StreamPass — История разработки

> Формат: Дата | Что сделано | Файлы | Результат

---

| Дата | Что сделано | Файлы | Результат |
|------|-------------|-------|-----------|
| 2026-08-03 | BL-001/BL-002: Hysteria2 tunnel + AAR | `client/go_core/`, `client/android/app/libs/streampasscore.aar`, Kotlin bridge | go_core transport, Gradle AAR dep, assembleDebug OK |
| 2026-08-03 | Инициализация AI-friendly документации | `docs/*`, `ai/*`, `reports/*`, `prompts/*` | Полная база знаний для AI-агентов |
| — | Первый релиз StreamPass MVP (backend) | `backend/`, `shared/`, `docker-compose.yml` | Backend API functional |
| — | Клиентская часть (Flutter Android) | `client/lib/`, `client/android/` | UI screens + VPNService scaffold |
| — | Связка backend с Android client | `client/lib/services/`, handlers | API integration working |
| — | Fix: real relay data in VPN service | `StreamPassVpnService.kt`, relay handler | Relay config passed to VPN |
| — | Health Monitor worker | `backend/cmd/healthmonitor/` | TCP probe + health reporting |
| — | Migration 0002: connection_config | `0002_relay_connection_config.up.sql` | Relay connection secrets in DB |
| — | Remove build artifacts from VCS | `.gitignore` | Cleaner repo |
| — | Update .gitignore | `.gitignore` | android_old, build dirs ignored |

---

## Git Commits (хронология)

| Commit | Message |
|--------|---------|
| `9570b95` | Первый релиз StreamPass MVP |
| `3c854d5` | Update .gitignore |
| `b78ba57` | Добавлена клиентская часть проекта |
| `98bcb63` | Добавление доработки |
| `df531de` | chore: remove build artifacts and debug backups from version control |
| `50e9e15` | fix: восстановлена передача реальных данных relay в VPN-сервис |
| `aae6501` | Добавление правок |
| `6a6f029` | Связка бэкенд с клинтом Андроид |

---

## Шаблон для новых записей

```
| YYYY-MM-DD | [Описание задачи] | [Изменённые файлы] | [Результат] |
```

Обновлять после каждой завершённой задачи.
