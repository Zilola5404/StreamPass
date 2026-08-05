#!/usr/bin/env bash
# Install / repair daily encrypted off-site backup (BL-035).
# Runs after the local Postgres dump cron (default 03:15 UTC).
#
# Usage on primary VPS (as root):
#   bash scripts/install-offsite-backup-cron.sh
#   OFFSITE_SSH=root@212.43.157.167:/var/backups/streampass bash scripts/install-offsite-backup-cron.sh
#
# Requires in /root/StreamPass/.env (or environment):
#   BACKUP_ENCRYPT_KEY=...   # long random secret
# Optional:
#   OFFSITE_DIR=/var/backups/streampass-offsite   # encrypted local mirror
#   OFFSITE_SSH=user@host:/path                  # true off-box copy via scp

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/streampass}"
OFFSITE_DIR="${OFFSITE_DIR:-/var/backups/streampass-offsite}"
OFFSITE_SSH="${OFFSITE_SSH:-}"
CRON_HOUR="${CRON_HOUR:-3}"
CRON_MIN="${CRON_MIN:-15}"
CRON_MARKER="# streampass-offsite-backup"
ENV_FILE="${ENV_FILE:-$ROOT/.env}"

chmod +x "$ROOT/scripts/backup-postgres.sh" \
  "$ROOT/scripts/backup-offsite.sh" \
  "$ROOT/scripts/restore-postgres.sh" \
  "$ROOT/scripts/install-backup-cron.sh" \
  "$ROOT/scripts/install-offsite-backup-cron.sh" 2>/dev/null || true

mkdir -p "$BACKUP_DIR" "$OFFSITE_DIR"
chmod 700 "$BACKUP_DIR" "$OFFSITE_DIR"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

if [[ -z "${BACKUP_ENCRYPT_KEY:-}" ]]; then
  echo "ERROR: BACKUP_ENCRYPT_KEY missing (set in $ENV_FILE)" >&2
  exit 1
fi

# Prefer env override for OFFSITE_SSH; else keep .env value.
if [[ -n "${OFFSITE_SSH}" ]]; then
  export OFFSITE_SSH
fi
export BACKUP_DIR OFFSITE_DIR BACKUP_ENCRYPT_KEY

echo "Running local Postgres dump..."
BACKUP_DIR="$BACKUP_DIR" RETENTION_DAYS="${RETENTION_DAYS:-30}" \
  "$ROOT/scripts/backup-postgres.sh"

echo "Running encrypted off-site sync..."
BACKUP_DIR="$BACKUP_DIR" OFFSITE_DIR="$OFFSITE_DIR" \
  BACKUP_ENCRYPT_KEY="$BACKUP_ENCRYPT_KEY" \
  OFFSITE_SSH="${OFFSITE_SSH:-}" \
  "$ROOT/scripts/backup-offsite.sh"

cron_line="${CRON_MIN} ${CRON_HOUR} * * * cd ${ROOT} && set -a && . ${ENV_FILE} && set +a && BACKUP_DIR=${BACKUP_DIR} OFFSITE_DIR=${OFFSITE_DIR} ${ROOT}/scripts/backup-offsite.sh >> ${BACKUP_DIR}/offsite.log 2>&1 ${CRON_MARKER}"

existing="$(crontab -l 2>/dev/null || true)"
filtered="$(printf '%s\n' "$existing" | grep -vF "$CRON_MARKER" || true)"
{
  printf '%s\n' "$filtered"
  echo "$cron_line"
} | crontab -

# Ensure postgres dump script stays executable in its cron too.
bash "$ROOT/scripts/install-backup-cron.sh"

echo "Installed off-site cron:"
crontab -l | grep -E 'streampass-(postgres|offsite)-backup' || true
echo "OFFSITE_DIR=$OFFSITE_DIR"
echo "OFFSITE_SSH=${OFFSITE_SSH:-'(none — local encrypted mirror only)'}"
