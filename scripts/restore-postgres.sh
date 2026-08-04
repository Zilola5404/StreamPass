#!/usr/bin/env bash
# StreamPass — restore PostgreSQL from gzip SQL dump (BL-033)
# Usage:
#   ./scripts/restore-postgres.sh /path/to/streampass_YYYYMMDD_HHMMSS.sql.gz
#   CONFIRM=yes ./scripts/restore-postgres.sh backups/streampass_latest.sql.gz
#
# Stops backend briefly to avoid writes during restore.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
DB_USER="${DB_USER:-streampass}"
DB_NAME="${DB_NAME:-streampass}"
CONFIRM="${CONFIRM:-}"

dump="${1:-}"
if [[ -z "$dump" ]]; then
  echo "Usage: $0 <backup.sql.gz>" >&2
  exit 2
fi
if [[ ! -f "$dump" ]]; then
  echo "ERROR: file not found: $dump" >&2
  exit 1
fi

if [[ "$CONFIRM" != "yes" ]]; then
  echo "This will REPLACE database '$DB_NAME'."
  echo "Re-run with CONFIRM=yes to proceed:"
  echo "  CONFIRM=yes $0 $dump"
  exit 3
fi

compose() {
  docker compose -f "$COMPOSE_FILE" --project-directory "$ROOT" "$@"
}

cid="$(compose ps -q postgres | head -n1)"
if [[ -z "$cid" ]]; then
  echo "ERROR: postgres container not running" >&2
  exit 1
fi

echo "Stopping backend + healthmonitor..."
compose stop backend healthmonitor >/dev/null

echo "Recreating empty database..."
docker exec -i "$cid" psql -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 <<SQL
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();
DROP DATABASE IF EXISTS $DB_NAME;
CREATE DATABASE $DB_NAME OWNER $DB_USER;
SQL

echo "Restoring $dump ..."
gzip -dc "$dump" | docker exec -i "$cid" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 >/dev/null

echo "Starting backend + healthmonitor..."
compose start backend healthmonitor >/dev/null

echo "OK: restore complete from $dump"
