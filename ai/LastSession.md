# Last Session

> Updated: 2026-08-06

## Completed

- Pushed **BL-017** TCP underlay (`44cf5e9`)
- **BL-035** off-site backups:
  - Fixed `Permission denied` on cron (`chmod +x`)
  - SSH key primary → secondary; encrypted scp to `212.43.157.167`
  - Cron 03:00 dump / 03:15 off-site
  - Operator pull via `PullBackupsOffsite.ps1` → `backups/offsite/`

## Residual

- Branded domain, ЮKassa, Windows/iOS/macOS — not started (excluded / needs user)
- Physical device connect still manual (no adb)
