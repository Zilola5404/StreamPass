# Code Review — BL-017 + BL-035

> **Commits:** `44cf5e9` (BL-017 TCP underlay), `9acdf8b` (BL-035 off-site backup)  
> **Дата:** 2026-08-06 | **Reviewer:** dev follow-up (QA M-05)

---

## Summary

Изменения изолированы: BL-017 — `client/go_core/internal/hyconfig` + VPS ops scripts; BL-035 — backup scripts + cron. Backend API не затронут. Риск регрессии низкий при green unit/integration tests.

---

## BL-017 — TCP underlay fallback

| Аспект | Оценка |
|--------|--------|
| Архитектура | ✅ Fallback candidates отделены; TCP underlay framing в отдельном пакете |
| Безопасность | ✅ Нет новых secrets в коде; bridge только на VPS |
| Тесты | ✅ `fallback_test.go`, `tcp_underlay_test.go`, integration test |
| Отклонение от ТЗ §10 | ⚠️ TCP/443 пропущен — задокументировано **ADR-014** |

**Рекомендация:** Accept с ADR-014.

---

## BL-035 — Off-site backup

| Аспект | Оценка |
|--------|--------|
| Шифрование | ✅ AES-256-CBC + PBKDF2 перед копированием |
| Секреты | ✅ `BACKUP_ENCRYPT_KEY` из env; `-pass env:` (не argv) |
| Ops | ✅ Cron 03:15 UTC; `VerifyOffsiteBackup.ps1` для prod check |
| Restore | ⚠️ Restore drill не автоматизирован — manual по `27_BackupRecovery.md` |

**Рекомендация:** Accept после prod verify (M-02).

---

## Verdict

**APPROVE** (code) — при условии green tests и prod verification off-site backup.
