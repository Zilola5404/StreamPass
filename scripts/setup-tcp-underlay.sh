#!/usr/bin/env bash
# Install StreamPass TCP underlay bridge on a relay host.
# Forwards framed TCP datagrams (ports 8443, 24443) → local Hysteria UDP/443.
#
# Usage on VPS:
#   bash scripts/setup-tcp-underlay.sh
# Or from repo root after copying this tree to the server.
set -euo pipefail

LISTEN="${LISTEN:-:8443,:24443}"
UDP_TARGET="${UDP_TARGET:-127.0.0.1:443}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
SRC_DIR="$(cd "$(dirname "$0")/tcpunderlay" && pwd)"

if ! command -v go >/dev/null 2>&1; then
  echo "go toolchain required to build tcpunderlay" >&2
  exit 1
fi

echo "Building tcpunderlay from ${SRC_DIR}..."
(cd "$SRC_DIR" && CGO_ENABLED=0 go build -o /tmp/streampass-tcpunderlay .)
install -m 0755 /tmp/streampass-tcpunderlay "${INSTALL_DIR}/streampass-tcpunderlay"

cat > /etc/systemd/system/streampass-tcpunderlay.service <<EOF
[Unit]
Description=StreamPass TCP underlay bridge (Hysteria2)
After=network.target hysteria.service
Wants=hysteria.service

[Service]
ExecStart=${INSTALL_DIR}/streampass-tcpunderlay -listen ${LISTEN} -udp ${UDP_TARGET}
Restart=always
RestartSec=3
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable streampass-tcpunderlay
systemctl restart streampass-tcpunderlay

if command -v ufw >/dev/null 2>&1; then
  ufw allow 8443/tcp || true
  ufw allow 24443/tcp || true
fi

sleep 1
systemctl status streampass-tcpunderlay --no-pager || true
ss -lntp | grep -E ':(8443|24443)\b' || true
echo "TCP underlay ready: TCP ${LISTEN} → UDP ${UDP_TARGET}"
