# Device traffic QA — StreamPass +42 (adb)

> **Date:** 2026-08-10 14:05–14:12 (UTC+3)  
> **Device:** SM-S938B (One UI / Android 16), adb `RFGYB48SWAF`  
> **App:** `com.streampass.app` versionName=0.1.1 versionCode=42  
> **Session:** `nl-native-1` `212.43.156.33:443`, `networkMode=split`, `ipv6=capture`, hysteria `pingMs=513`

---

## Verdict

**FAIL — Connected, but sites/apps do not load.**

VPN tunnel is up (`tun0`, key icon in status bar), Hysteria handshake OK, UI Connected — **browser/apps cannot resolve DNS and cannot open TCP to destination IPs.**

---

## Automated checks (adb)

| Check | Result | Detail |
|-------|--------|--------|
| `tun0` up | PASS | `10.10.0.1/30` MTU 1400 |
| Private DNS mode | PASS | `off` (cleared specifier during test) |
| Hysteria / PrepareRelay | PASS | `udp/443` pingMs=513, `nl-native-1` |
| Resolve `www.google.com` | **FAIL** | `curl: (6) Could not resolve host` |
| HTTPS Google by IP | **FAIL** | `curl: (28) Connection timed out` (8s) |
| HTTPS `1.1.1.1` trace | **FAIL** | timeout |
| Chrome `ya.ru` | **FAIL** | screenshot: «Не удалось найти IP-адрес сервера ya.ru» |
| YouTube DNS (netd) | **FAIL** | `isBlocked=true` |
| WhatsApp via Go DoH | PARTIAL | `[dns] query … via=doh` works when query hits VPN DNS |

Screenshot: `reports/QA/streampass-connected-screen.png`  
Log extract: `reports/QA/connect-logcat-adb-2026-08-10.txt`

---

## Root signals

1. **System DNS not consistently hitting VPN DNS (`10.10.0.1`)**  
   - Log: `[dns] warn host_empty ip=57.144…` — apps dial IP literals without HostForIP.  
   - Chrome fails hostname lookup for `ya.ru` / Google.

2. **VPN network not validated by Android**  
   - `network{291}` Score: `EVER_EVALUATED&IS_UNMETERED&IS_VPN` — **no `IS_VALIDATED` / `EVER_VALIDATED`**.  
   - Wi‑Fi remains the validated Internet network.

3. **RELAY path reports `result=ok` with `latency_ms=0 speed_kbps=0`** for Meta/Google IPs — dial may open without useful payload.

4. **ICMP pings** to `8.8.8.8` / VPS show ~0.5 ms — **not trustworthy** (local/fake); ignore for SLA.

---

## Connect timing (this session)

| Phase | Time |
|-------|------|
| session begin → hysteria ok | ~0.53 s |
| TUN establish (11406 RU excludes) | ~4.0 s |
| begin → `tunnel event=connected` | **~4.58 s** (**PASS ≤5s**) |

---

## Next product fix / retest

1. Force all app DNS to VPN `10.10.0.1` (block Private DNS / Chrome Secure DNS / bypass).  
2. Confirm `relay-tcp rewrite` appears for Google/YouTube after HostForIP works.  
3. Retest **Full Relay** (no 11k excludes) vs Split.  
4. Make VPN network **VALIDATED** (Android NetworkMonitor probe through tunnel).
