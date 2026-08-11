# Current Task

> Updated: 2026-08-10

## Shipping: +49

### Bug (adb on +48)
`2ip.ru` DNS→DIRECT but TCP `host=` empty → `default_relay_foreign` (Android A cache bypasses VPN DNS).

### Fix
On AAAA-suppress: prefetch TypeA + `PinDirectIP` / `PinRelayIP` before answering.

### Expect
- ifconfig/ipify → NL `212.43.156.33`
- 2ip TCP → DIRECT (ISP), logs `pin-direct` + `mode=DIRECT`
