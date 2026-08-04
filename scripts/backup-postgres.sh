#!/usr/bin/env bash
# StreamPass — automated PostgreSQL backup (BL-033)
# Usage:
#   ./scripts/backup-postgres.sh
#   BACKUP_DIR=/var/backups/streampass RETENTION_DAYS=30 ./scripts/backup-postgres.sh
#
# Designed for Linux production (cron). Reads DB credentials from compose .env
# via the running postgres container (no password on the CLI).

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-$ROOT/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
DB_USER="${DB_USER:-streampass}"
DB_NAME="${DB_NAME:-streampass}"
CONTAINER="${POSTGRES_CONTAINER:-}"

mkdir -p "$BACKUP_DIR"

timestamp="$(date -u +%Y%m%d_%H%M%S)"
outfile="$BACKUP_DIR/streampass_${timestamp}.sql.gz"
logfile="$BACKUP_DIR/backup.log"

log() {
  local msg="[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*"
  echo "$msg" | tee -a "$logfile"
}

compose() {
  docker compose -f "$COMPOSE_FILE" --project-directory "$ROOT" "$@"
}

resolve_container() {
  if [[ -n "$CONTAINER" ]]; then
    echo "$CONTAINER"
    return
  fi
  # Prefer compose service name; fall back to first running postgres container.
  if compose ps -q postgres >/dev/null 2>&1; then
    local id
    id="$(compose ps -q postgres | head -n1)"
    if [[ -n "$id" ]]; then
      echo "$id"
      return
    fi
  fi
  docker ps --filter "ancestor=postgres:16-alpine" --format '{{.ID}}' | head -n1
}

cid="$(resolve_container)"
if [[ -z "$cid" ]]; then
  log "ERROR: postgres container not found"
  exit 1
fi

log "Starting backup → $outfile (container=$cid)"
if ! docker exec -i "$cid" pg_dump -U "$DB_USER" -d "$DB_NAME" --no-owner --no-acl | gzip -c >"$outfile.tmp"; then
  rm -f "$outfile.tmp"
  log "ERROR: pg_dump failed"
  exit 1
fi
mv "$outfile.tmp" "$outfile"

size_kb="$(du -k "$outfile" | awk '{print $1}')"
if [[ "$size_kb" -lt 1 ]]; then
  log "ERROR: backup file empty"
  rm -f "$outfile"
  exit 1
fi

# Integrity: gzip must be valid and dump must look like SQL
if ! gzip -t "$outfile"; then
  log "ERROR: gzip integrity check failed"
  exit 1
fi
if ! gzip -dc "$outfile" | head -n 5 | grep -qE '^(--|SET |CREATE |PostgreSQL)'; then
  log "WARN: dump header did not match expected SQL markers"
fi

log "OK: $outfile (${size_kb} KB)"

# Retention
deleted=0
while IFS= read -r -d '' old; do
  rm -f "$old"
  deleted=$((deleted + 1))
done < <(find "$BACKUP_DIR" -maxdepth 1 -type f -name 'streampass_*.sql.gz' -mtime +"$RETENTION_DAYS" -print0 2>/dev/null || true)
log "Retention: kept last ${RETENTION_DAYS}d, removed $deleted old file(s)"

# Latest symlink for ops convenience
ln -sfn "$(basename "$outfile")" "$BACKUP_DIR/streampass_latest.sql.gz"
log "Latest → $BACKUP_DIR/streampass_latest.sql.gz"
