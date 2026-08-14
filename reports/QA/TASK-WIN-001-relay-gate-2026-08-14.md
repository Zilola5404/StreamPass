# TASK-WIN-001 — Independent Hysteria2 → Native Relay gate

> Date: 2026-08-14  
> Client: official Hysteria2 Go library (`github.com/apernet/hysteria/core/v2/client`) via `go_core/internal/hyconfig` integration tests  
> Relay: production `nl-native-1` `212.43.156.33:443` (healthy)  
> URI: **not logged** (fetched at runtime from `GET /api/v1/servers`)

This is **not** the StreamPass TUN / Android VPN path. Handshake-only is not accepted as “internet works”.

## Results

| Check | Target | Result |
|-------|--------|--------|
| Handshake | native relay | **PASS** (0.55 s) |
| TCP + first byte | `ifconfig.me:80` | **PASS** — public IPv4 `212.43.156.33` (relay exit = native VPS) |
| HTTP | `example.com:80` HEAD | **PASS** |
| HTTPS / TLS | `example.com:443` | **PASS** `HTTP/1.1 200 OK` |
| HTTPS / TLS | `github.com:443` | **PASS** `HTTP/1.1 200 OK` |
| HTTPS / TLS | `www.google.com:443` | **PASS** `HTTP/1.1 200 OK` |
| HTTPS / TLS | `www.youtube.com:443` | **PASS** `HTTP/1.1 200 OK` |
| DNS over UDP | `example.com` A via `1.1.1.1:53` | **PASS** (61 bytes) |

Command (secrets via env only):

```text
go test -count=1 -timeout 3m -v -run TestIntegrationHysteria(Connect|ForeignIP|UDPEcho|HTTPHead|HTTPS)$ ./internal/hyconfig/
```

`ok streampass/go_core/internal/hyconfig  5.692s`

## Verdict

**Native Relay is usable. TASK-WIN-001 Windows Smart Routing development may proceed** from step 2 (skeleton), then Auth/API, then TUN — not from full-tunnel-all-traffic.

## Notes

- Hiddify `nl-amsterdam-1` was not in the healthy `/servers` list this run; gate used **native** `nl-native-1`.
- Production URIs currently include `insecure=1` (self-signed). TZ forbids `insecure=1` in production Windows client; TLS pin/cert must be fixed before Windows production ship.
- Exit IP equals the relay host (single NL VPS). Expected until regional exit nodes exist.
