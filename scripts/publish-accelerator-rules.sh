#!/usr/bin/env bash
# Publish accelerator rule set (ТЗ §6 / docs/07.4): RU DIRECT, foreign media RELAY.
# Domain-first; no Cloudflare /12. Usage: ADMIN_API_KEY=... bash scripts/publish-accelerator-rules.sh

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

    {"kind":"DOMAIN","pattern":"gemini.google.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.gemini.google.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"ai.google.dev","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"aistudio.google.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"generativelanguage.googleapis.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.googleapis.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"google.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.google.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"chatgpt.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.chatgpt.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"openai.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.openai.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"claude.ai","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.claude.ai","mode":"RELAY"},

    {"kind":"DOMAIN","pattern":"telegram.org","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.telegram.org","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"t.me","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.t.me","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"linkedin.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.linkedin.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"licdn.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.licdn.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"upwork.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.upwork.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"indeed.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.indeed.com","mode":"RELAY"},

    {"kind":"DOMAIN","pattern":"github.com","mode":"RELAY"},
    {"kind":"DOMAIN","pattern":"*.github.com","mode":"RELAY"},

    {"kind":"CIDR","pattern":"91.108.4.0/22","mode":"RELAY"},
    {"kind":"CIDR","pattern":"91.108.8.0/22","mode":"RELAY"},
    {"kind":"CIDR","pattern":"91.108.56.0/22","mode":"RELAY"},
    {"kind":"CIDR","pattern":"149.154.160.0/20","mode":"RELAY"}
  ]
}
JSON

echo "POST $API_BASE/rules"
curl -fsS -X POST "$API_BASE/rules" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: $ADMIN_API_KEY" \
  --data-binary @/tmp/accelerator_rules.json | python3 -m json.tool
