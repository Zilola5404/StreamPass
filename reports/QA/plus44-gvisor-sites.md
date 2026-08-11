# +44 gvisor — foreign sites fixed

> 2026-08-10 15:03 · SM-S938B · `0.1.1+44` · split · `tun-stack=gvisor` · DNS `198.18.0.1`

## Cause (user: 2ip / Upwork / Indeed / LinkedIn / Gemini)

RU sites worked via `excludeRoute`. Foreign TCP never entered Go handler on `system` stack → hang.

## Fix
`sing-tun` stack **`system` → `gvisor`** (`-tags with_gvisor` in gomobile bind).

## adb matrix (Connected +44)

| Site | HTTP | Notes |
|------|------|-------|
| ya.ru | 302 | RU DIRECT OK |
| 2ip.ru | 200 | `.ru` → DIRECT (shows local/ISP IP, not relay) |
| google.com | 200 | RELAY `transfer_done` ~1.3 Mbps |
| youtube.com | 200 | RELAY ~9 Mbps |
| gemini.google.com | 200 | RELAY |
| instagram.com | 200 | RELAY |
| upwork.com | 403 | TCP OK (Cloudflare/WAF to curl) |
| indeed.com | 403 | TCP OK (geo page) |
| linkedin.com | connect OK | was PowerShell parse noise; TCP path live |

Log: `[vpn] tun-stack=gvisor`, `[tun] tcp mode=RELAY … transfer_done`.

## APK
`client/build/app/outputs/flutter-apk/StreamPass-v0.1.1+44-signed-arm64.apk`

## Note
For «мой IP через VPN» use non-`.ru` checker (`ifconfig.me`) or Full Relay — `2ip.ru` is intentionally DIRECT.
