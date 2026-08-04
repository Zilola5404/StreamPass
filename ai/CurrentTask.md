# Current Task

> Updated: 2026-08-04

## Done

- **BL-026** multi-region: catalog `de/nl/pl/fi`, `GET /regions`, `?region=` filter,
  normalize on register, client picker, Admin region select, migration `0005`
- **BL-027** `go.sum` already tracked in repo → marked Done
- Admin login UX fix (session no longer cleared on health-check failure)

## Ops note

Production still has **NL-only** relays. To add Frankfurt/Warsaw/Helsinki:
1. Provision VPS + Hysteria
2. Admin → Relays → register with region `de` / `pl` / `fi`

## Next

- Device E2E +16 / new region APK
- BL-023 Windows / BL-030 billing / BL-033 backups
