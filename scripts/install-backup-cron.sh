#!/usr/bin/env bash
# Install daily cron for StreamPass Postgres backups (BL-033).
# Usage (as root on the VPS):
#   bash scripts/install-backup-cron.sh
#   BACKUP_DIR=/var/backups/streampass RETENTION_DAYS=30 bash scripts/install-backup-cron.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/streampass}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
CRON_HOUR="${CRON_HOUR:-3}"
CRON_MARKER="# streampass-postgres-backup"

mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"
chmod +x "$ROOT/scripts/backup-postgres.sh" "$ROOT/scripts/restore-postgres.sh"

# Smoke once now
BACKUP_DIR="$BACKUP_DIR" RETENTION_DAYS="$RETENTION_DAYS" "$ROOT/scripts/backup-postgres.sh"

cron_line="0 ${CRON_HOUR} * * * BACKUP_DIR=${BACKUP_DIR} RETENTION_DAYS=${RETENTION_DAYS} ${ROOT}/scripts/backup-postgres.sh >> ${BACKUP_DIR}/cron.log 2>&1 ${CRON_MARKER}"

# Replace previous streampass backup cron entries
existing="$(crontab -l 2>/dev/null || true)"
filtered="$(printf '%s\n' "$existing" | grep -vF "$CRON_MARKER" || true)"
{
  printf '%s\n' "$filtered"
  echo "$cron_line"
} | crontab -

echo "Installed cron:"
crontab -l | grep -F "$CRON_MARKER" || true
echo "Backups: $BACKUP_DIR (retention ${RETENTION_DAYS}d, daily ${CRON_HOUR}:00 UTC)"
