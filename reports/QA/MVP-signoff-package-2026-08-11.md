# MVP Sign-Off Package — 2026-08-11

> Client: **v0.1.1+54** installed (`versionCode=2054`, `com.streampass.app`)  
> Backend: `9560c60` on `https://212-43-156-33.nip.io` (migrations 0009–0010)  
> Device: Samsung SM-S938B (One UI)

---

## Evidence checklist (for PO / TL / QA)

| Area | Result | Evidence |
|------|--------|----------|
| Auth register/login/refresh | ✅ | API + onboarding |
| Devices / limit (BL-049) | ✅ | Profile: **Активных: 2 из 3**, «Android (это устройство)» after re-login |
| Failure notifications UI (BL-051) | ✅ | Settings → Оформление |
| Admin Premium / ban / audit (BL-050) | ✅ | API smoke earlier + `/admin/` |
| Routing split (RU DIRECT / foreign RELAY) | ✅ | Prior +50/+52 adb |
| Instagram feed | ✅ | +52 must-relay blackhole fix |
| ЮKassa live | ❌ Blocked | Need shop/secret keys (BL-040) |
| Theme / About / password eye (BL-052+) | ✅ | APK +54 on device; Settings → тема / О приложении / глаз пароля |

---

## SLA retest (BL-053) — 2026-08-11 evening

| Metric | Measured | Budget | Status |
|--------|----------|--------|--------|
| API `/health` | 0.63 s | ≤ 2 s | **PASS** |
| API `/config` | 0.27 s | ≤ 2 s | **PASS** |
| Cold `am start` TotalTime | 133–149 ms | ≤ 2 s | **PASS** |
| Hysteria handshake `pingMs` | **382 ms** | ≤ 5 s | **PASS** |
| Connect session begin → tunnel connected | **~6.04 s** (34.618→40.657) | ≤ 5 s | **FAIL** (+1.0 s) — TUN establish on One UI split |
| Recover after airplane | Attempted via `adb cmd connectivity airplane-mode`; UI interrupted by system shade; **not a clean formal PASS** | ≤ 10 s | **PENDING** manual (toggle airplane in-app Connected state, stopwatch) |

Log: `reports/QA/bl053-connect-logcat.txt`

---

## Sign-off table

| Role | Name | Date | Decision | Signature |
|------|------|------|----------|-----------|
| Product Owner | | 2026-08-11 | Accept MVP **with** BL-040 exception / Reject | |
| Tech Lead | | 2026-08-11 | Accept / Reject (note T2 TUN) | |
| QA | | 2026-08-11 | Accept / Reject | |

**Recommended PO decision:** Accept Android MVP functionally **except** live payments (BL-040) and note T2 soft-miss on Samsung split TUN (~6 s).

---

## Residual blockers

1. **BL-040** ЮKassa live keys  
2. Optional: T2 TUN optimization; formal T3 airplane stopwatch  
3. T4 30d uptime (ops)
