# QA Response — PRODUCT-QA-2026-08-07

> **Дата ответа:** 2026-08-07  
> **На отчёт:** `reports/QA/PRODUCT-QA-2026-08-07.md`  
> **Код:** `routing-policy-v1` + Google/Meta CIDR IP-only safety net → client **v0.1.1+34**

## Вердикт по блокерам

| ID | Действие разработчика | Статус |
|----|----------------------|--------|
| BUG-001 | IP-only Meta/Google без HostForIP снова → **RELAY** (CIDR safety net; Cloudflare `/12` не возвращали). Rules API republish. | **FIXED in +34 / rules** — нужен device retest |
| BUG-002 | App bypass packages уже включают `ru.rostel` + heuristics; после OTA+reconnect проверить log `VPN app-bypass: …` | **Mitigated / verify on device** |
| BUG-003 | OTA APK + config `latest_client_version=0.1.1+34`, docs `03_CurrentState` | **FIXED** (deploy this session) |
| BUG-004 | Network error ≠ inactive subscription (`_subscriptionCheckFailed`) | **Already in tree** — QA retest |
| BUG-005 | BL-035 off-site | **OPEN** (out of this traffic fix) |

## Что установить на device

1. APK OTA: `https://212-43-156-33.nip.io/downloads/StreamPass.apk`  
2. Или локально: `client/build/app/outputs/flutter-apk/StreamPass-v0.1.1+34-signed-arm64.apk`  
3. Private DNS = Off; Network Mode = Split; reconnect VPN  
4. Expect in connect.log: `vpn dns=10.10.0.1`, `build=0.1.1+34`, `[decision] … action=RELAY` for YouTube/Meta IPs

## Retest checklist (from QA §17)

```powershell
.\scripts\DiagnoseTrafficBlock.ps1 -LiveProbe -ReportPath reports\QA\traffic-block-retest.md
.\scripts\VerifyAppSiteSwitch.ps1 -AutoLaunch -SkipManual -ReportPath reports\QA\traffic-switch-retest.md
```

Pass: YouTube/Instagram/Gemini open; Gosuslugi without VPN block; no empty-host DIRECT for 142.250/157.240.
