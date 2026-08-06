#!/usr/bin/env bash
# Encrypt latest Postgres dump and copy off-box (BL off-site follow-up).
# Usage on VPS:
#   BACKUP_ENCRYPT_KEY=... OFFSITE_SSH=user@host:/path bash scripts/backup-offsite.sh
#   # or local second path:
#   OFFSITE_DIR=/var/backups/streampass-offsite bash scripts/backup-offsite.sh
#
# Always encrypts before leaving the primary backup directory.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/streampass}"
OFFSITE_DIR="${OFFSITE_DIR:-}"
OFFSITE_SSH="${OFFSITE_SSH:-}"
KEY="${BACKUP_ENCRYPT_KEY:-}"
LOG="${BACKUP_DIR}/offsite.log"

log() { echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*" | tee -a "$LOG"; }

if [[ -z "$KEY" ]]; then
  log "ERROR: set BACKUP_ENCRYPT_KEY (long random string)"
  exit 1
fi
if [[ -z "$OFFSITE_DIR" && -z "$OFFSITE_SSH" ]]; then
  log "ERROR: set OFFSITE_DIR and/or OFFSITE_SSH"
  exit 1
fi

src="${BACKUP_DIR}/streampass_latest.sql.gz"
if [[ ! -f "$src" ]]; then
  log "ERROR: missing $src — run backup-postgres.sh first"
  exit 1
fi

ts="$(date -u +%Y%m%d_%H%M%S)"
enc="/tmp/streampass_${ts}.sql.gz.enc"
export BACKUP_ENCRYPT_KEY="$KEY"
openssl enc -aes-256-cbc -pbkdf2 -salt -pass env:BACKUP_ENCRYPT_KEY -in "$src" -out "$enc"
log "Encrypted $(du -h "$enc" | awk '{print $1}')"

if [[ -n "$OFFSITE_DIR" ]]; then
  mkdir -p "$OFFSITE_DIR"
  cp -f "$enc" "$OFFSITE_DIR/streampass_${ts}.sql.gz.enc"
  ln -sfn "streampass_${ts}.sql.gz.enc" "$OFFSITE_DIR/streampass_latest.sql.gz.enc"
  # retention 30d
  find "$OFFSITE_DIR" -maxdepth 1 -type f -name 'streampass_*.sql.gz.enc' -mtime +30 -delete || true
  log "Copied to OFFSITE_DIR=$OFFSITE_DIR"
fi

if [[ -n "$OFFSITE_SSH" ]]; then
  # OFFSITE_SSH is user@host:/dir — ensure remote directory exists when possible.
  remote_host="${OFFSITE_SSH%%:*}"
  remote_path="${OFFSITE_SSH#*:}"
  if [[ -n "$remote_host" && -n "$remote_path" && "$remote_path" != "$OFFSITE_SSH" ]]; then
    ssh -o StrictHostKeyChecking=accept-new -o BatchMode=yes "$remote_host" \
      "mkdir -p '$remote_path' && chmod 700 '$remote_path'" || \
      log "WARN: could not mkdir on remote (will still try scp)"
  fi
  scp -o StrictHostKeyChecking=accept-new "$enc" "$OFFSITE_SSH/streampass_${ts}.sql.gz.enc"
  # Keep a stable latest name on the remote when ssh works.
  ssh -o StrictHostKeyChecking=accept-new -o BatchMode=yes "$remote_host" \
    "ln -sfn 'streampass_${ts}.sql.gz.enc' '$remote_path/streampass_latest.sql.gz.enc'" 2>/dev/null || true
  log "Uploaded to OFFSITE_SSH=$OFFSITE_SSH"
fi

rm -f "$enc"
log "OK offsite sync"
