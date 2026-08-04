# StreamPass — Backup & Recovery

> Дата: 2026-08-04 | BL-033 Done

---

## Current State

| Component | Backup | Status |
|-----------|--------|--------|
| PostgreSQL | Daily cron + gzip dumps | ✅ Automated (BL-033) |
| Redis | No persistence (no save/AOF) | Sessions ephemeral |
| Caddy certs | Docker volume `caddy_data` | Auto-renewed by Caddy |
| Code | Git repository | ✅ Remote on origin/main |
| .env secrets | Manual | ⚠️ Not in git (by design) |

---

## PostgreSQL Backup

### Production (Linux VPS)

```bash
# One-shot
bash scripts/backup-postgres.sh

# Install daily cron (03:00 UTC, 30-day retention)
bash scripts/install-backup-cron.sh

# Custom location / retention
BACKUP_DIR=/var/backups/streampass RETENTION_DAYS=30 bash scripts/install-backup-cron.sh
```

Artifacts:
- `$BACKUP_DIR/streampass_YYYYMMDD_HHMMSS.sql.gz`
- `$BACKUP_DIR/streampass_latest.sql.gz` (symlink)
- `$BACKUP_DIR/backup.log`

### Local / Windows

```powershell
.\scripts\Backup.ps1
.\scripts\Backup.ps1 -OutputDir .\backups -RetentionDays 30
```

### Restore

```bash
# Dry gate — requires explicit confirmation
CONFIRM=yes bash scripts/restore-postgres.sh /var/backups/streampass/streampass_latest.sql.gz
```

Restore stops `backend` + `healthmonitor`, recreates the DB, loads the dump, then starts them again.

---

## Redis Recovery

Redis configured with `--save ""` and `--appendonly no`.  
**Impact:** All sessions lost on Redis restart. Users must re-login.  
**Acceptable for MVP.** For production: enable AOF or RDB.

---

## Disaster Recovery

### Scenario: VPS failure

1. Provision new VPS
2. Install Docker + Docker Compose
3. Clone repo, restore `.env`
4. Restore PostgreSQL: `CONFIRM=yes bash scripts/restore-postgres.sh <dump.sql.gz>`
5. `docker compose up -d --build`
6. Update DNS if IP changed
7. Verify `/health`
8. Re-install cron: `bash scripts/install-backup-cron.sh`

**RTO:** ~30–60 min (new VPS + restore)  
**RPO:** ≤ 24 h (daily backup at 03:00 UTC)

### Scenario: Database corruption

1. `CONFIRM=yes bash scripts/restore-postgres.sh /var/backups/streampass/streampass_latest.sql.gz`
2. Verify Admin → Users / Relays
3. Spot-check `/health`

### Scenario: Migration failure

1. Check `schema_migrations` table
2. Run corresponding `.down.sql` manually
3. Fix migration, redeploy

---

## Secrets Recovery

`.env` file is NOT in git.  
**Store securely:** password manager, encrypted backup, secret manager.  
Required secrets: `DB_PASSWORD`, `JWT_SECRET`, `ADMIN_API_KEY`, `YOOKASSA_*`

---

## Client Data

| Data | Storage | Backup |
|------|---------|--------|
| Auth tokens | SharedPreferences (device) | User re-login |
| Settings | SharedPreferences + native | Lost on uninstall |
| Exclusions | Local + synced (BL-014) | Server is source of truth |

---

## Recommendations (next)

1. Off-site copy of `$BACKUP_DIR` (S3 / second VPS) — not yet automated
2. Monthly restore drill
3. Enable Redis persistence when session durability matters
4. Encrypted off-box copy of `.env`
