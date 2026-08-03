# Current Focus

> Updated: 2026-08-03

## Главный фокус проекта сейчас

**E2E VPN test на Android (BL-003) — подтверждение MVP tunnel.**

Hysteria2 transport в go_core реализован, AAR подключён. Следующий критический шаг — проверка на устройстве с live relay.

## Why This Focus

- Backend: MVP functional
- Android UI: screens + VPNService готовы
- VPN tunnel: go_core + AAR — done (BL-001/002)
- Decision Engine / Rule Engine — следующий этап после E2E

## What NOT to focus on now

- Admin Panel UI (P2)
- iOS/Windows/macOS clients (post-MVP)
- Prometheus/Grafana (post-MVP)
- Крупные рефакторинги документации без изменений кода

## Success Metric

User presses Connect → VPN active → ifconfig.me shows relay IP.

## Related Backlog

BL-003 → BL-005 → BL-006 → BL-004 → BL-010
