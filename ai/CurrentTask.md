# Current Task

> Updated: 2026-08-12

## Done (Android MVP close-out)
- PO/TL/QA sign-off (MVP without live YooKassa; TUN ~6s noted)
- +54 password eye committed/pushed
- **BL-040** → Telegram Stars + USDT per TZ 02.3 (code + `/pay/`; bot token pending from PO)
- **BL-052** theme/about re-ship without SpAppBar (+55)
- Commit/push `d2dcdb6` on `main`
- APK built: `downloads/StreamPass-v0.1.1+55-signed-arm64.apk` (12.9 MB)

## Blocked right now
- **VPS deploy** from this PC: `212.43.156.33` ports 22/443 timeout (SSH/HTTPS unreachable)
- When network is back, on VPS:
  ```bash
  cd /root/StreamPass && git pull --ff-only origin main
  docker compose up -d --build backend caddy
  ```
  Then upload APK to `/root/StreamPass/downloads/` and set `TELEGRAM_BOT_TOKEN` when PO provides it.

## Next
- **Windows client (BL-023)** — next epic
- After PO provides `TELEGRAM_BOT_TOKEN`: set env, restart backend, smoke `/buy` + Stars pay
- Optional: T3 recover stopwatch; TUN ≤5s optimization

## Note
- test5 password was reset for adb smoke to `StreamPass53a` — change if needed
- No adb device attached during +55 build (install locally when phone is connected)
