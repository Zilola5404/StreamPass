# +43 partial traffic — why some sites work

> 2026-08-10 14:40–14:54 · SM-S938B · `0.1.1+43` · `nl-native-1` · **split** · DNS `198.18.0.1` · iface `tun1`

## Verdict

| Destination | Path | Result |
|-------------|------|--------|
| ya.ru, vk.com, mail.ru, gosuslugi | **OS DIRECT** (RU CIDR `excludeRoute`) | **PASS** (~0.2s) |
| google.com, youtube, instagram, 1.1.1.1 | **TUN → RELAY** | **FAIL** (TCP connect timeout) |
| connectivitycheck.gstatic.com | **TUN → DIRECT** (rule) | **FAIL** (same TCP timeout) |

DNS on +43 works (`198.18.0.1`, DoH/Yandex). Sites fail **after** resolve: TCP into TUN never reaches `[tun] tcp …` handler.

## Evidence

```
build=0.1.1+43  vpn dns=198.18.0.1  networkMode=split  exclude-ru 11406
hysteria ok pingMs=171  traffic_ready via=RELAY (once)
ya.ru → HTTP 302 ~0.15s
www.google.com → DNS OK → Trying 142.251.x:443 → timeout 6–8s
  (no [tun] tcp / relay-tcp lines during curl)
connectivitycheck.gstatic.com → DNS DIRECT → TCP :80 timeout
VPN Score: IS_VPN only — **not IS_VALIDATED**
```

## Root cause (simple)

1. **Russian sites open** because split mode throws their IPs **out of the VPN** onto Wi‑Fi — Go/Hysteria never see that TCP.
2. **Foreign sites** must go **through TUN**. UDP DNS inside TUN works; **userspace TCP (`sing-tun` stack=`system`) does not complete connections** on this One UI 8 / Android 16 device — curl SYNs time out with zero Go TCP logs.
3. So: not “random sites”, but **DIRECT-bypass vs TUN path**. Anything that needs the tunnel (Google/YouTube/Instagram, and even NetworkMonitor) breaks.

## Next fix (+44)

- Switch TUN stack `system` → **`gvisor`** (common fix when system stack drops TCP on OEM Android).
- Retest Full Relay + Split; confirm `[tun] tcp mode=RELAY` / `transfer_done` and VPN `IS_VALIDATED`.
