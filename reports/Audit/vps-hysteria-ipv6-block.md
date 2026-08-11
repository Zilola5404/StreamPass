# VPS Hysteria journal — IPv6 block (2026-08-09)

> Host: `212.43.156.33` · Unit: `hysteria` · Client seen: `94.77.165.47`

## Verdict

**Client ↔ relay path works.** Packets reach `hysteria`; the process dials targets. Dominant failure: client asks relay to open **IPv6** destinations on an **IPv4-only** VPS → `no IPv4 address available`.

## Evidence (`journalctl -u hysteria`)

Sample errors (client `94.77.165.47`):

| reqAddr | error |
|---------|--------|
| `[2a03:2880:…]:443` (Meta) | `no IPv4 address available` |
| `[2606:4700::…]:443` (Cloudflare) | `no IPv4 address available` |
| `[2a00:1450:…]:443` (Google) | `no IPv4 address available` |
| `[2a02:6b8:…]:443` (Yandex) | `no IPv4 address available` |
| `142.251.39.131:443` (IPv4) | dial attempted; some timeouts / Application error 0x0 |
| `212.43.157.167:443` | `no route to host` (hairpin/wrong target) |

`systemctl is-active hysteria` → **active**. Disconnects after flood of failed dials.

## Why Instagram “works” more often

Many IG CDN endpoints resolve to **IPv4** first; Happy Eyeballs / AAAA-preferring apps (Meta, Google, CF, Ya) fail hard on AAAA → Hysteria.

## Fix (client **v0.1.1+40**)

1. DNS: answer **AAAA** with empty NOERROR (`aaaa-suppress`) — force A / IPv4.
2. TUN: drop native IPv6 destinations early (`[tun] drop ipv6 … relay IPv4-only`).
3. Keep Private DNS Off / DoT :853 blocked so VPN DNS is used.

## QA after +40

1. Install APK, Private DNS = Off, Connect split.
2. Expect device: `[dns] … via=aaaa-suppress`, few/no IPv6 `reqAddr` on VPS.
3. Optional live: `journalctl -u hysteria -f` while browsing — should see IPv4 `reqAddr` dials, not `[2a…]/`/`[2606…]`.
