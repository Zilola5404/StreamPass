#!/usr/bin/env bash
# Publish StreamPass client config (min version, auto-update fields).
# Usage:
#   ADMIN_API_KEY=... bash scripts/publish-client-config.sh
#   API_BASE=https://212-43-156-33.nip.io/api/v1 MIN_SUPPORTED=0.1.0 LATEST=0.1.1 bash scripts/publish-client-config.sh

set -euo pipefail

API_BASE="${API_BASE:-https://212-43-156-33.nip.io/api/v1}"
ADMIN_API_KEY="${ADMIN_API_KEY:?Set ADMIN_API_KEY}"
MIN_SUPPORTED="${MIN_SUPPORTED:-0.1.0}"
LATEST="${LATEST:-0.1.1}"
DOWNLOAD_URL="${DOWNLOAD_URL:-}"
RULE_POLL="${RULE_POLL:-300}"
RELAY_POLL="${RELAY_POLL:-60}"

body="$(printf '{"min_supported_client_version":"%s","latest_client_version":"%s","client_download_url":"%s","telemetry_enabled":true,"rule_poll_interval_sec":%s,"relay_poll_interval_sec":%s}' \
  "$MIN_SUPPORTED" "$LATEST" "$DOWNLOAD_URL" "$RULE_POLL" "$RELAY_POLL")"

echo "POST $API_BASE/config min=$MIN_SUPPORTED latest=$LATEST"
curl -fsS -X POST "$API_BASE/config" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: $ADMIN_API_KEY" \
  -d "$body" | python3 -m json.tool

echo
curl -fsS "$API_BASE/config" | python3 -c "import sys,json; c=json.load(sys.stdin); print('live version', c['version'], 'min', c['min_supported_client_version'], 'latest', c.get('latest_client_version',''))"
