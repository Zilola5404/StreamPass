#!/usr/bin/env bash
# StreamPass — установка Hysteria2 relay на Ubuntu/Debian (по docs/02_TZ.md)
# Запуск на сервере: bash setup-relay-hysteria.sh

set -euo pipefail

HYSTERIA_VERSION="${HYSTERIA_VERSION:-v2.6.1}"
AUTH_PASSWORD="${AUTH_PASSWORD:-streampass-secure-auth}"
OBFS_PASSWORD="${OBFS_PASSWORD:-streampass-relay-2024}"
LISTEN_PORT="${LISTEN_PORT:-443}"

echo "=== StreamPass Hysteria2 relay setup ==="

if [[ $EUID -ne 0 ]]; then
  echo "Run as root: sudo bash $0"
  exit 1
fi

apt-get update -qq
apt-get install -y -qq curl openssl ca-certificates

ARCH=$(uname -m)
case "$ARCH" in
  x86_64) HY_ARCH=amd64 ;;
  aarch64) HY_ARCH=arm64 ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

URL="https://github.com/apernet/hysteria/releases/download/app/${HYSTERIA_VERSION}/hysteria-linux-${HY_ARCH}"
echo "Downloading $URL"
curl -fsSL "$URL" -o /usr/local/bin/hysteria
chmod +x /usr/local/bin/hysteria
/usr/local/bin/hysteria version

mkdir -p /etc/hysteria
if [[ ! -f /etc/hysteria/cert.pem ]]; then
  openssl req -x509 -nodes -newkey rsa:2048 \
    -keyout /etc/hysteria/key.pem \
    -out /etc/hysteria/cert.pem \
    -days 3650 -subj "/CN=relay.streampass"
fi

cat > /etc/hysteria/config.yaml <<EOF
listen: :${LISTEN_PORT}

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

cat > /etc/systemd/system/hysteria.service <<'EOF'
[Unit]
Description=Hysteria2 Relay (StreamPass)
After=network.target

[Service]
ExecStart=/usr/local/bin/hysteria server -c /etc/hysteria/config.yaml
Restart=always
RestartSec=5
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable hysteria
systemctl restart hysteria

if command -v ufw >/dev/null 2>&1; then
  ufw allow "${LISTEN_PORT}/udp" || true
  ufw allow "${LISTEN_PORT}/tcp" || true
fi

sleep 2
systemctl status hysteria --no-pager || true
ss -lunp | grep ":${LISTEN_PORT}" || true

PUBLIC_IP=$(curl -fsSL -4 ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')
echo ""
echo "=== Done ==="
echo "connection_config:"
echo "hysteria2://${AUTH_PASSWORD}@${PUBLIC_IP}:${LISTEN_PORT}/?obfs=salamander&obfs-password=${OBFS_PASSWORD}#StreamPass"
