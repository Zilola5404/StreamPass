# Last Session

> Updated: 2026-09-10

## Completed

- RELEASE-001 production readiness package on `release/mvp-1`:
  - Reset `traffic_ready` on `SetHysteriaClient` / reconnect
  - Android underlay recover: CONNECTING until new first_byte
  - Canonical `[lifecycle]` CONNECT/RELAY/TRAFFIC_READY/RECONNECT events
  - DefaultMode SSOT docs = DIRECT (ADR-020 supersedes ADR-018 runtime)
  - Diagnostics IPv6 platform-accurate strings
  - Reports: `RELEASE-001-Production-Readiness.md`

## Residual (blocks 100/100)

- Physical Android + Windows E2E evidence on this tip
- Real Platega payment E2E
- Rebuild AAR with JDK (`gomobile bind -tags with_gvisor`)
- Live Backend OFF / Relay OFF blackhole tests on devices
- Do **not** declare Public Beta until Device matrix PASS
