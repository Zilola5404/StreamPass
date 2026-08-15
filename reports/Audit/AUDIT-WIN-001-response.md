# AUDIT-WIN-001 — ответ на замечания архитектора

**Дата:** 2026-08-15  
**Коммит-основа:** c4d2f6d (аудит)  
**Статус:** Engineering Validation (не Production Ready)

---

## Закрытые пункты (код + документация)

| # | Замечание аудита | Действие | Статус |
|---|------------------|----------|--------|
| P0-5 | WFP vs Wintun | **ADR-020:** Wintun для MVP; WFP отложен. Обновлены TZ, Roadmap, Architecture, SaaS scenarios | ✅ Doc |
| P0-6 | DefaultMode DIRECT vs RELAY | SSOT: **`07.4` + ADR-018 = RELAY**. Синхронизированы FS §6, CurrentState, MasterPrompt, DirectVsVpnBypass | ✅ Doc |
| P0-7 | `insecure=1` в production | Sidecar **блокирует** connect без `STREAMPASS_ALLOW_INSECURE=1`; deploy scripts требуют `RELAY_PIN_SHA256` | ✅ Code |
| P1-8 | Хрупкий route cleanup `10.10.0.2` | Session gateway registry + scan `route print` для 10.10.0.x; cleanup on stop | ✅ Code |
| P1-9 | Underlay по эвристике | Приоритет **OS default-route NIC** (`Get-NetRoute`), затем dial-owner/probe | ✅ Code |
| §10 | Ложный единый PASS | `VerifyWindowsTUN.ps1` и **новый** `VerifyWindowsE2E.ps1` разделяют Unit / Integration / Live / Browser E2E | ✅ Scripts |
| Deploy | Нет Windows deploy doc | `docs/26_Deployment.md` — секция Windows Client Build | ✅ Doc |

---

## Открытые пункты (требуют физического Windows E2E)

| # | Пункт DoD | Статус |
|---|-----------|--------|
| Live Wintun Admin | `VerifyWindowsTUN.ps1` live section | ⏳ Manual |
| DIRECT browser E2E | ya.ru, 2ip, gosuslugi, s7 | ⏳ `VerifyWindowsE2E.ps1` |
| RELAY browser E2E | YouTube, GitHub, Instagram | ⏳ Manual |
| DNS + HostForIP E2E | `host=` не пустой систематически | ⏳ Manual |
| Relay-down | DIRECT RU при падении relay | ⏳ Manual |
| Reconnect / stale route | После disconnect нет blackhole | ⏳ Manual (session route cleanup добавлен) |

---

## Как прогнать проверки

```powershell
# Automated layers (Admin for live TUN)
powershell -NoProfile -ExecutionPolicy Bypass -File scripts\VerifyWindowsTUN.ps1

# E2E checklist + report template
powershell -NoProfile -ExecutionPolicy Bypass -File scripts\VerifyWindowsE2E.ps1 -ReportPath reports\QA\win-e2e-latest.md
```

---

## Вердикт

**AUDIT-WIN-001 остаётся открытым** по data-plane E2E, но архитектурные блокеры (WFP/Wintun, DefaultMode, insecure, route/underlay) **закрыты в репозитории**. Следующий шаг — physical browser E2E на Windows 10/11, не расширение CIDR/rules.
