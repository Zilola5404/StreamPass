# StreamPass — Backup & Recovery

> Дата: 2026-08-03

---

## Current State

| Component | Backup | Status |
|-----------|--------|--------|
| PostgreSQL | Docker volume `postgres_data` | ⚠️ No automated backup |
| Redis | No persistence (no save/AOF) | Sessions ephemeral |
| Caddy certs | Docker volume `caddy_data` | Auto-renewed by Caddy |
| Code | Git repository | ✅ Remote on origin/main |
| .env secrets | Manual | ⚠️ Not in git (by design) |

---

## PostgreSQL Backup

### Manual Backup
```bash
# From host
docker compose exec postgres pg_dump -U streampass streampass > backup_$(date +%Y%m%d).sql

# Restore
docker compose exec -T postgres psql -U streampass streampass < backup_20260803.sql
```

### Automated Backup (TODO: BL-033)
```bash
# Planned cron (daily)
0 3 * * * docker compose exec -T postgres pg_dump -U streampass streampass | gzip > /backups/streampass_$(date +\%Y\%m\%d).sql.gz
```

**Script:** `scripts/Backup.ps1` — TODO: implement

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
4. Restore PostgreSQL from latest backup
5. `docker compose up -d --build`
6. Update DNS if IP changed
7. Verify `/health`

**RTO:** TODO: Требуется уточнение (depends on backup frequency)  
**RPO:** TODO: Требуется уточнение (depends on backup frequency)

### Scenario: Database corruption

1. Stop backend: `docker compose stop backend`
2. Restore from backup (see above)
3. Restart: `docker compose start backend`
4. Verify data integrity

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
| Exclusions | Local only | Lost on uninstall |

No server-side device backup.

---

## Recommendations (Pre-Production)

1. Daily automated PostgreSQL backup with 30-day retention
2. Off-site backup storage (S3, another VPS)
3. Test restore procedure monthly
4. Document RTO/RPO targets
5. Enable Redis persistence for production
6. Backup `.env` securely outside repo
