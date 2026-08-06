# QA Response — BL-017 + BL-035

> **Дата:** 2026-08-06 | Ответ разработчика на `reports/QA/BL-017-BL-035-QA.md`

---

## Закрыто в этом проходе

| ID | Статус | Fix |
|----|--------|-----|
| **C-01** | ✅ | `TestDecideRoute_defaultRelay` → `TestDecideRoute_defaultDirect` (ожидает `DIRECT` = `decision.DefaultMode`) |
| **C-02** | ✅ | `connect_flow_log_test.dart`: `TestWidgetsFlutterBinding.ensureInitialized()`; `VpnChannel.connect` валидирует config до EventChannel |
| **M-01** | ✅ | Документация синхронизирована на **v0.1.1+25** (`03_CurrentState`, `14_AIContext`, `99_ProjectDashboard`, …) |
| **M-04** | ✅ | **ADR-014**: пропуск TCP/443 (Caddy), TCP underlay 8443/24443 |
| **m-01** | ✅ | Лог `[connect] hysteria ok via udp/443` / `tcp/8443` в `mobile/tunnel.go` → diagnostics |
| **m-02** | ✅ | `docs/10_Progress.md` — записи BL-017/035 |
| **m-03** | ✅ | `docs/06_TestPlan.md` — дата прогона 2026-08-06, 49/49 flutter |
| **I-01** | ✅ | ADR-014 документирует TCP/24443 |
| **I-02** | ✅ | `backup-offsite.sh`: `-pass env:BACKUP_ENCRYPT_KEY` |
| **I-03** | ✅ | `PullBackupsOffsite.ps1`: SSH key (`STREAMPASS_SSH_KEY`) с fallback на password |

## Добавлено

- `scripts/VerifyOffsiteBackup.ps1` — проверка cron + `offsite.log` + `.enc` на secondary (требует SSH)

## Проверки

```
client/go_core: go test ./...  — PASS (incl. mobile)
client:         flutter test   — 49/49 PASS
backend:        go test ./...  — PASS
```

## Остаётся (ops / manual QA)

| ID | Статус | Действие |
|----|--------|----------|
| **M-02** | ⏳ | Запустить `.\scripts\VerifyOffsiteBackup.ps1` с SSH к VPS |
| **M-03** | ⏳ | Manual connect на устройстве +25 (Connect → YouTube / `.ru` DIRECT) |
| **M-05** | ℹ️ | Code review — см. `reports/CodeReview/BL-017-BL-035-review.md` |

---

*После M-02 + M-03 QA может перевести вердикт в PASS.*
