package api

import (
	"github.com/gin-gonic/gin"
	"github.com/aipo/server/internal/api/handler"
	"github.com/aipo/server/internal/api/middleware"
	"github.com/aipo/server/internal/db"
	"github.com/aipo/server/internal/hub"
)

const uninstallScript = `#!/bin/bash
set -e
systemctl stop aipo-agent 2>/dev/null || true
systemctl disable aipo-agent 2>/dev/null || true
rm -f /etc/systemd/system/aipo-agent.service
systemctl daemon-reload
rm -rf /opt/aipo-agent
echo "Agent uninstalled."
`

const installScript = `#!/bin/bash
set -e

SERVER_URL="${AIPO_SERVER_URL:-ws://localhost:8080/ws/agent}"
ENROLL_TOKEN="${AIPO_ENROLL_TOKEN:?AIPO_ENROLL_TOKEN is required}"

INSTALL_DIR="/opt/aipo-agent"
mkdir -p "$INSTALL_DIR"

ARCH=$(uname -m)
case $ARCH in x86_64) ARCH="amd64";; aarch64) ARCH="arm64";; esac
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

HTTP_BASE=$(echo "$SERVER_URL" | sed 's|^ws://|http://|;s|^wss://|https://|' | sed 's|/ws/agent||')

echo "Downloading agent binary..."
curl -fsSL "$HTTP_BASE/releases/agent-${OS}-${ARCH}" -o "$INSTALL_DIR/agent"
chmod +x "$INSTALL_DIR/agent"

cat > "$INSTALL_DIR/agent.yaml" <<EOF
server_url: $SERVER_URL
enroll_token: $ENROLL_TOKEN
EOF
chmod 600 "$INSTALL_DIR/agent.yaml"

cat > /etc/systemd/system/aipo-agent.service <<EOF
[Unit]
Description=aipo Agent
After=network.target

[Service]
ExecStart=$INSTALL_DIR/agent -config $INSTALL_DIR/agent.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now aipo-agent
echo "Agent installed and started."
`

func NewRouter(store db.Store, h *hub.Hub, jwtSecret, adminUser, adminPass, enrollToken string) *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	authH := handler.NewAuthHandler(store, jwtSecret, adminUser, adminPass)
	agentH := handler.NewAgentHandler(store, h, enrollToken)
	agentWS := handler.NewAgentWSHandler(store, h, enrollToken)
	termH := handler.NewTerminalHandler(h)

	// Agent install script & binaries (public)
	r.GET("/install.sh", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(200, installScript)
	})
	r.GET("/uninstall.sh", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(200, uninstallScript)
	})
	r.Static("/releases", "./releases")

	// Agent WebSocket endpoint (no JWT — uses token in register message)
	r.GET("/ws/agent", agentWS.Connect)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", authH.Login)

		authed := v1.Group("/", middleware.Auth(jwtSecret))
		{
		fileH := handler.NewFileHandler(h)

		authed.GET("/agents", agentH.List)
			authed.DELETE("/agents/:id", agentH.Delete)
			authed.GET("/agents/enroll-token", agentH.EnrollToken)
			authed.GET("/agents/:id/terminal", termH.Connect)
			authed.GET("/agents/:id/files", fileH.List)
			authed.GET("/agents/:id/files/download", fileH.Download)
			authed.POST("/agents/:id/files/upload", fileH.Upload)
			authed.DELETE("/agents/:id/files", fileH.Delete)
		}
	}

	return r
}
