# StreamPass — Master Prompt

> Дата: 2026-08-06 | Для старта любой AI-сессии

Ты работаешь над **StreamPass** — Go backend + Flutter Android **ускоритель трафика** (не full-tunnel VPN) с умной маршрутизацией.

1. Прочитай `docs/00_ProjectRules.md` и `docs/14_AIContext.md`.  
2. Статус задач — только из `docs/04_Backlog.md` (не из старых reports).  
3. Текущий фокус — `ai/CurrentTask.md`.  
4. DIRECT vs VPN-флаг ОС — `docs/33_DirectVsVpnBypass.md`.  

**Факты (не устарели):**
- Клиент `v0.1.1+34`: Hysteria2 + TUN + Decision (`DefaultMode=DIRECT`) + DNS-in-TUN + HostForIP + RU CIDR split-tunnel + app-bypass + TCP underlay (`routing-policy-v1`).  
- Domain DIRECT **не** снимает `TRANSPORT_VPN` — для Госуслуг/банков нужен `addDisallowedApplication`.  
- Prod: `https://212-43-156-33.nip.io`, Admin `/admin/`.  
- Done: CI, Admin, monitoring, regions, backups, auto-update, E2E mock, loadtest.  
- Не трогать без запроса: ЮKassa live, Windows/iOS/macOS.  

API: `/api/v1/*`, JWT для клиента, `X-Admin-Key` для admin.
