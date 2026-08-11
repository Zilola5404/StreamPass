# +43 ship + Full Relay QA status

> 2026-08-10

## Built & installed
- APK: `StreamPass-v0.1.1+43-signed-arm64.apk` (versionCode=43)
- Device SM-S938B: `adb install -r` Success

## Code changes
| Item | Change |
|------|--------|
| DNS | `addDnsServer(198.18.0.1)` instead of `10.10.0.1` |
| TCP DNS | Go hijacks TCP/53 like UDP/53 |
| Full Relay | Home Connect honors Diagnostics networkMode |
| Validation | NetworkMonitor hosts → fast Yandex/UDP DNS first |
| Log label | PackageInfo version (was hardcoded +42) |

## Full Relay device pass
**Not completed via adb UI** — automation kept landing on Settings; no new `tun0` / `networkMode=full_relay` session observed.

**Manual steps for user:**
1. Настройки → Диагностика → **Full Relay**
2. Главная → **nl-native-1** → Connect
3. Confirm Connected; open google.com / ya.ru
4. Reply when Connected — agent will re-run curl/DNS/validation checks
