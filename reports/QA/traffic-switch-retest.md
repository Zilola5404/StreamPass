# Traffic switch QA report

Generated: 2026-08-08 00:07

| ID | Label | Type | Expected | Status | Problems |
|----|-------|------|----------|--------|----------|
| site_yandex | Yandex (RU) | site | DIRECT | WARN | /W/A/R/N/:/ /e/x/p/e/c/t/e/d/ /s/i/g/n/ /'/R/U/ /C/I/D/R/ /b/y/p/a/s/s/ /а/к/т/и/в/е/н/ /(/и/з/ /c/o/n/n/e/c/t/)/'/ /n/o/t/ /f/o/u/n/d/ /i/n/ /l/o/g/s/ |
| site_2ip | 2ip.ru (geo check) | site | DIRECT | PASS | /-/ |
| site_youtube | YouTube (foreign) | site | RELAY | WARN | /W/A/R/N/:/ /e/x/p/e/c/t/e/d/ /s/i/g/n/ /'/F/o/r/e/i/g/n/ /D/N/S/ /ч/е/р/е/з/ /D/o/H/'/ /n/o/t/ /f/o/u/n/d/ /i/n/ /l/o/g/s/;/ /W/A/R/N/:/ /e/x/p/e/c/t/e/d/ /s/i/g/n/ /'/R/e/l/a/y/ /п/о/д/к/л/ю/ч/ё/н/'/ /n/o/t/ /f/o/u/n/d/ /i/n/ /l/o/g/s/ |
| site_instagram | Instagram (foreign) | site | RELAY | WARN | /W/A/R/N/:/ /e/x/p/e/c/t/e/d/ /s/i/g/n/ /'/R/e/l/a/y/ /c/o/n/n/e/c/t/e/d/'/ /n/o/t/ /f/o/u/n/d/ /i/n/ /l/o/g/s/ |
| site_gemini | Gemini (Google AI) | site | RELAY | WARN | /W/A/R/N/:/ /e/x/p/e/c/t/e/d/ /s/i/g/n/ /'/G/e/m/i/n/i/ /D/N/S/ /v/i/a/ /D/o/H/'/ /n/o/t/ /f/o/u/n/d/ /i/n/ /l/o/g/s/;/ /W/A/R/N/:/ /e/x/p/e/c/t/e/d/ /s/i/g/n/ /'/R/e/l/a/y/ /c/o/n/n/e/c/t/e/d/'/ /n/o/t/ /f/o/u/n/d/ /i/n/ /l/o/g/s/ |
| app_gosuslugi | Госуслуги | app | BYPASS | PASS | /-/ |
| app_sber | Сбербанк | app | BYPASS | WARN | /W/A/R/N/:/ /e/x/p/e/c/t/e/d/ /s/i/g/n/ /'/С/б/е/р/ /в/ /b/y/p/a/s/s/-/л/и/с/т/е/'/ /n/o/t/ /f/o/u/n/d/ /i/n/ /l/o/g/s/ |
| app_streampass_api | StreamPass (self API) | app | BYPASS | WARN | /W/A/R/N/:/ /e/x/p/e/c/t/e/d/ /s/i/g/n/ /'/S/e/l/f/-/b/y/p/a/s/s/:/ /A/P/I/ /н/е/ /ч/е/р/е/з/ /h/a/i/r/p/i/n/'/ /n/o/t/ /f/o/u/n/d/ /i/n/ /l/o/g/s/;/ /W/A/R/N/:/ /e/x/p/e/c/t/e/d/ /s/i/g/n/ /'/A/P/I/ /о/т/в/е/т/и/л/'/ /n/o/t/ /f/o/u/n/d/ /i/n/ /l/o/g/s/ |
| switch_chrome_back | Chrome ↔ StreamPass | lifecycle |  | PASS | /-/ |

## Legend
- **site** - Chrome URL, check DIRECT/RELAY + DNS in logs
- **app** - native app, check VPN bypass + no VPN block dialog
- **lifecycle** - task switch stability (StreamPass must stay alive)