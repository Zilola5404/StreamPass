#!/usr/bin/env bash
# Publish accelerator rule set (ТЗ §6): all Russian TLDs DIRECT, foreign media RELAY.
# Usage: ADMIN_API_KEY=... bash scripts/publish-accelerator-rules.sh

set -euo pipefail

API_BASE="${API_BASE:-https://212-43-156-33.nip.io/api/v1}"
ADMIN_API_KEY="${ADMIN_API_KEY:?Set ADMIN_API_KEY}"

cat >/tmp/accelerator_rules.json <<'JSON'
{
  "rules": [
    {"kind":"DOMAIN","pattern":"*.ru","mode":"DIRECT"},
    {"kind":"DOMAIN","pattern":"*.su","mode":"DIRECT"},
    {"kind":"DOMAIN","pattern":"*.xn--p1ai","mode":"DIRECT"},
    {"kind":"DOMAIN","pattern":"*.mos.ru","mode":"DIRECT"},
    {"kind":"DOMAIN","pattern":"*.gov.ru","mode":"DIRECT"},
    {"kind":"DOMAIN","pattern":"*.edu.ru","mode":"DIRECT"},
    {"kind":"DOMAIN","pattern":"*.mil.ru","mode":"DIRECT"},

    {"kind":"DOMAIN","pattern":"instagram.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.instagram.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"cdninstagram.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.cdninstagram.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"facebook.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.facebook.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"fbcdn.net","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.fbcdn.net","mode":"RELAY"},

    {"kind":"DOMAIN","pattern":"youtube.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.youtube.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"youtu.be","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"ytimg.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.ytimg.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"googlevideo.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.googlevideo.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"github.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.github.com","mode":"RELAY"}
  ]
}
JSON

echo "POST $API_BASE/rules"
curl -fsS -X POST "$API_BASE/rules" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: $ADMIN_API_KEY" \
  --data-binary @/tmp/accelerator_rules.json | python3 -m json.tool
