# Next Task

> 2026-08-15

TASK-WIN-001 stages 5-12 automated verify is in repo.

**Operator:** run as Administrator:
```
powershell -NoProfile -ExecutionPolicy Bypass -File scripts\VerifyWindowsTUN.ps1
```
Then manual checklist in the generated `reports/QA/TASK-WIN-001-stages5-12-*.md`.

**Before production ship:** remove `insecure=1` (TLS pin).
