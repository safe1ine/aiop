#!/bin/bash
set -e

SERVER_URL="${1:-ws://localhost:8080/ws/agent}"
AGENT_NAME="${2:-$(hostname)}"
AGENT_TOKEN="${3:-}"

if [ -z "$AGENT_TOKEN" ]; then
  echo "Usage: $0 <server_url> <agent_name> <agent_token>"
  echo "Example: $0 ws://192.168.1.100:8080/ws/agent my-server abc123..."
  exit 1
fi

INSTALL_DIR="/opt/aipo-agent"
mkdir -p "$INSTALL_DIR"

# Download binary (adjust URL for your server)
ARCH=$(uname -m)
case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
esac
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

echo "Downloading agent binary..."
# curl -fsSL "https://your-server/releases/agent-${OS}-${ARCH}" -o "$INSTALL_DIR/agent"
# For now, assume binary is already built and placed here
chmod +x "$INSTALL_DIR/agent"

# Write config
cat > "$INSTALL_DIR/agent.yaml" <<EOF
server_url: $SERVER_URL
name: $AGENT_NAME
token: $AGENT_TOKEN
EOF
chmod 600 "$INSTALL_DIR/agent.yaml"

# Install systemd service
cat > /etc/systemd/system/aipo-agent.service <<EOF
[Unit]
Description=aipo Agent
After=network.target

[Service]
ExecStart=$INSTALL_DIR/agent -config $INSTALL_DIR/agent.yaml
Restart=always
RestartSec=5
WorkingDirectory=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable aipo-agent
systemctl start aipo-agent

echo "Agent installed and started."
systemctl status aipo-agent --no-pager
