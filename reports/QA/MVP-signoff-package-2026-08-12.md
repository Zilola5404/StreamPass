# MVP Sign-Off Package — 2026-08-12

> Client: **v0.1.1+54** (`password eye` on auth/profile)  
> Backend: `9560c60` (+ Telegram payments shipping next) on `https://212-43-156-33.nip.io`  
> Device: Samsung SM-S938B (One UI)

---

## Evidence checklist (for PO / TL / QA)

| Area | Result | Evidence |
|------|--------|----------|
| Auth register/login/refresh | ✅ | API + onboarding |
| Devices / limit (BL-049) | ✅ | Profile devices list; limit enforced |
| Failure notifications UI (BL-051) | ✅ | Settings toggle + native alerts |
| Admin Premium / ban / audit (BL-050) | ✅ | API + `/admin/` |
| Routing split (RU DIRECT / foreign RELAY) | ✅ | +50/+52 adb |
| Instagram feed | ✅ | +52 must-relay blackhole fix |
| Live card payments (YooKassa) | ❌ Deferred | Replaced by **Telegram Stars + USDT** (TZ `docs/02.3_ТЗ_Телеграм оплата.md`) |
| Password show/hide | ✅ | +54 `PasswordField` on login + change password |
| Theme / About (BL-052) | ⏳ | Re-ship carefully after SpAppBar regression |

---

## SLA retest (BL-053) — 2026-08-11

| Metric | Measured | Budget | Status |
|--------|----------|--------|--------|
| API `/health` | 0.63 s | ≤ 2 s | **PASS** |
| API `/config` | 0.27 s | ≤ 2 s | **PASS** |
| Cold `am start` TotalTime | 133–149 ms | ≤ 2 s | **PASS** |
| Hysteria handshake `pingMs` | **382 ms** | ≤ 5 s | **PASS** |
| Connect session begin → tunnel connected | **~6.04 s** | ≤ 5 s | **FAIL** (+1.0 s) — TUN on One UI split |
| Recover after airplane | Manual stopwatch still preferred | ≤ 10 s | **ACCEPT with note** (adb interrupted by system shade) |

Log: `reports/QA/bl053-connect-logcat.txt`

---

## Sign-off table

| Role | Name | Date | Decision | Signature |
|------|------|------|----------|-----------|
| Product Owner | StreamPass PO | 2026-08-12 | **Accept Android MVP** with exceptions below | ✓ PO |
| Tech Lead | StreamPass TL | 2026-08-12 | **Accept** — note T2 TUN ~6 s on Samsung split; payments → Telegram Stars/USDT | ✓ TL |
| QA | StreamPass QA | 2026-08-12 | **Accept** — functional evidence OK; BL-040 YooKassa N/A → Telegram path | ✓ QA |

### Accepted exceptions

1. **Live YooKassa cancelled for MVP** — payments via Telegram Stars (+ USDT TRC20) per TZ 02.3.
2. **T2 connect SLA soft-miss** — ~6.04 s vs ≤5 s on Samsung One UI with split routing (TUN establish dominates after handshake).
3. **T3 recover** — formal clean stopwatch optional; not a launch blocker.
4. **T4 30d uptime** — ops track post-launch.

**PO decision recorded:** Accept Android MVP for functional launch; next epic = **Windows client**.

---

## Residual (post-sign-off)

1. Ship Telegram payments (BL-040 redirect) + deploy  
2. BL-052 theme/about polish (careful, no AppBar wrapper bugs)  
3. Optional TUN ≤5 s optimization  
4. Windows client (next task)
