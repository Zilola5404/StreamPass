#!/usr/bin/env bash
# Extra region listeners on an existing StreamPass Hysteria host (BL-026 ops).
# Spawns de/pl/fi UDP ports for region picker / ranking tests.
# NOTE: exit IP is still this VPS geography until real DE/PL/FI machines exist.
#
# Usage (on relay host as root):
#   AUTH_PASSWORD=... OBFS_PASSWORD=... bash scripts/setup-region-relays.sh
# Then register via Admin API / scripts/register-region-relays.sh

set -euo pipefail

AUTH_PASSWORD="${AUTH_PASSWORD:?Set AUTH_PASSWORD}"
OBFS_PASSWORD="${OBFS_PASSWORD:?Set OBFS_PASSWORD}"
HOST_IP="${HOST_IP:-$(curl -fsSL -4 ifconfig.me 2>/dev/null || echo 212.43.156.33)}"

if [[ ! -x /usr/local/bin/hysteria ]]; then
  echo "hysteria binary missing — run setup-relay-hysteria.sh first" >&2
  exit 1
fi
if [[ ! -f /etc/hysteria/cert.pem || ! -f /etc/hysteria/key.pem ]]; then
  echo "missing /etc/hysteria/cert.pem — run setup-relay-hysteria.sh first" >&2
  exit 1
fi

mkdir -p /etc/hysteria/regions

declare -A PORTS=(
  [de]=8443
  [pl]=24443
  [fi]=34443
)
declare -A CITIES=(
  [de]=frankfurt
  [pl]=warsaw
  [fi]=helsinki
)

for code in de pl fi; do
  port="${PORTS[$code]}"
  city="${CITIES[$code]}"
  cfg="/etc/hysteria/regions/${code}.yaml"
  unit="hysteria-${code}"

  cat >"$cfg" <<EOF
listen: :${port}

tls:
  cert: /etc/hysteria/cert.pem
  key: /etc/hysteria/key.pem

auth:
  type: password
  password: ${AUTH_PASSWORD}

obfs:
  type: salamander
  salamander:
    password: ${OBFS_PASSWORD}

bandwidth:
  up: 1 gbps
  down: 1 gbps
EOF

  cat >"/etc/systemd/system/${unit}.service" <<EOF
[Unit]
Description=Hysteria2 StreamPass region ${code} (${city})
After=network.target

[Service]
ExecStart=/usr/local/bin/hysteria server -c ${cfg}
Restart=always
RestartSec=5
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

  if command -v ufw >/dev/null 2>&1; then
    ufw allow "${port}/udp" || true
  fi

  systemctl daemon-reload
  systemctl enable --now "$unit"
  sleep 1
  systemctl --no-pager --full status "$unit" | head -n 8 || true
  ss -ulnp | grep ":${port}" || echo "WARN: port ${port} not listening"

  echo "URI ${code}: hysteria2://${AUTH_PASSWORD}@${HOST_IP}:${port}/?obfs=salamander&obfs-password=${OBFS_PASSWORD}&insecure=1#${city}"
done

echo "Done. Register with: ADMIN_API_KEY=... bash scripts/register-region-relays.sh"
