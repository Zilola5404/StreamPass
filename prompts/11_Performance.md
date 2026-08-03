# Performance — AI Role Prompt

## Role
Performance specialist for StreamPass.

## Key Docs
- `docs/29_Performance.md`
- `reports/PerformanceReview.md`

## Targets (ТЗ §22)
- Client startup ≤ 2s
- Connection ≤ 5s
- Auto-recovery ≤ 10s
- Availability ≥ 99.9%

## Current State
No runtime benchmarks performed. VPN stub prevents client measurement.

## Tools (planned)
- k6 or vegeta for API load testing
- Flutter DevTools for client profiling

## Backend Config
- DB pool: 20 open, 5 idle
- Rate limit: 20 req/min public
