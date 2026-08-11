# Relay vs Hiddify — 2026-08-09

## Verdict

Relay **alive**. VPS itself opens Google/Gmail/YouTube/Instagram (`curl` HTTP 200).
Hiddify **connects**, then fails mostly because client sends **IPv6** destinations to an **IPv4-only** outbound (`mode: 4`) → `no IPv4 address available` → browser `ERR_CONNECTION_CLOSED`.

## Evidence (Hiddify client `178.155.4.3`, ~45m)

| Metric | Count |
|--------|------:|
| Log lines for client | ~2031 |
| IPv6 `no IPv4 address available` | ~1919 |
| Other TCP errors | ~99 (timeout / reset after disconnect) |

Top targets: Yandex/CF/Google **AAAA** literals.

## What users must do in Hiddify/Windows

1. Prefer IPv4 only (disable IPv6 on adapter or Hiddify `ipv4_only` / domain strategy).
2. Or use DNS that does not return AAAA.

## Optional server follow-up

Enable Hysteria `protocolSniff` so TLS SNI → domain → resolve A with `mode: 4` (helps when client still sends IPv6 IP).
