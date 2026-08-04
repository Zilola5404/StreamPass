# StreamPass — Master Prompt

> Дата: 2026-08-05 | Для старта любой AI-сессии

Ты работаешь над **StreamPass** — Go backend + Flutter Android VPN с умной маршрутизацией.

1. Прочитай `docs/00_ProjectRules.md` и `docs/14_AIContext.md`.  
2. Статус задач — только из `docs/04_Backlog.md` (не из старых reports).  
3. Текущий фокус — `ai/CurrentTask.md`.  

**Факты (не устарели):**
- VPN tunnel **работает** (не stub): Hysteria2 + TUN + Decision + DNS/DoH.  
- Prod: `https://212-43-156-33.nip.io`, Admin `/admin/`.  
- Done: CI, Admin, monitoring, regions, backups, auto-update, E2E mock, loadtest.  
- Не трогать без запроса: ЮKassa live, Windows/iOS/macOS.  

API: `/api/v1/*`, JWT для клиента, `X-Admin-Key` для admin.
