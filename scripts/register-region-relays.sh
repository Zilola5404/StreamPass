#!/usr/bin/env bash
# Register de/pl/fi region relays against StreamPass Admin API.
# Usage:
#   ADMIN_API_KEY=... RELAY_PIN_SHA256=<64hex> bash scripts/register-region-relays.sh
#   RELAY_TLS=insecure_dev ADMIN_API_KEY=... bash scripts/register-region-relays.sh  # dev only
#   API_BASE=https://212-43-156-33.nip.io/api/v1 HOST_IP=212.43.156.33 bash scripts/register-region-relays.sh

set -euo pipefail

API_BASE="${API_BASE:-https://212-43-156-33.nip.io/api/v1}"
ADMIN_API_KEY="${ADMIN_API_KEY:?Set ADMIN_API_KEY}"
HOST_IP="${HOST_IP:-212.43.156.33}"
AUTH_PASSWORD="${AUTH_PASSWORD:?Set AUTH_PASSWORD}"
OBFS_PASSWORD="${OBFS_PASSWORD:?Set OBFS_PASSWORD}"
RELAY_TLS="${RELAY_TLS:-pin}"

build_uri() {
  local port="$1"
  local qs="obfs=salamander&obfs-password=${OBFS_PASSWORD}"
  if [[ "$RELAY_TLS" == "insecure_dev" ]]; then
    qs="${qs}&insecure=1"
  elif [[ -n "${RELAY_PIN_SHA256:-}" ]]; then
    qs="${qs}&pinSHA256=${RELAY_PIN_SHA256}&sni=${HOST_IP}"
  else
    echo "ERROR: set RELAY_PIN_SHA256 (production) or RELAY_TLS=insecure_dev (dev only)" >&2
    exit 1
  fi
  printf 'hysteria2://%s@%s:%s/?%s' "$AUTH_PASSWORD" "$HOST_IP" "$port" "$qs"
}

register() {
  local id="$1" region="$2" port="$3"
  local uri
  uri="$(build_uri "$port")"
  local body
  body="$(printf '{"id":"%s","region":"%s","host":"%s","port":%s,"connection_config":"%s"}' \
    "$id" "$region" "$HOST_IP" "$port" "$uri")"
  echo "Register $id ($region:$port) tls=$RELAY_TLS..."
  curl -fsS -X POST "$API_BASE/servers" \
    -H "Content-Type: application/json" \
    -H "X-Admin-Key: $ADMIN_API_KEY" \
    -d "$body" >/dev/null
  echo "OK $id"
}

register "de-frankfurt-1" "de" 8443
register "pl-warsaw-1" "pl" 24443
register "fi-helsinki-1" "fi" 34443

echo "All region relays registered."
curl -fsS -H "X-Admin-Key: $ADMIN_API_KEY" "$API_BASE/servers/all" | head -c 800
echo
