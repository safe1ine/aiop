package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/aipo/server/internal/db"
	"github.com/aipo/server/internal/hub"
	"github.com/aipo/server/internal/proto"
)

var agentUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type AgentWSHandler struct {
	store       db.Store
	hub         *hub.Hub
	enrollToken string
}

func NewAgentWSHandler(store db.Store, h *hub.Hub, enrollToken string) *AgentWSHandler {
	return &AgentWSHandler{store: store, hub: h, enrollToken: enrollToken}
}

func (h *AgentWSHandler) Connect(c *gin.Context) {
	conn, err := agentUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// First message must be register
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		return
	}
	conn.SetReadDeadline(time.Time{})

	var env proto.Envelope
	if err := json.Unmarshal(data, &env); err != nil || env.Type != proto.MsgRegister {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4001, "expected register"))
		return
	}

	var reg proto.RegisterPayload
	if err := json.Unmarshal(env.Payload, &reg); err != nil {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4001, "bad register payload"))
		return
	}

	if reg.EnrollToken != h.enrollToken {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4003, "invalid enroll token"))
		return
	}

	// Get client IP if agent didn't provide one
	ip := reg.IP
	if ip == "" {
		ip = c.ClientIP()
	}

	// Auto-register or update agent
	agent, err := h.store.UpsertAgent(reg.Hostname, ip, reg.OS, reg.Arch)
	if err != nil {
		log.Printf("upsert agent error: %v", err)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4000, "server error"))
		return
	}

	ac := h.hub.Register(agent.ID, agent.Hostname, conn)
	defer func() {
		h.hub.Unregister(agent.ID)
		h.store.UpdateAgentStatus(agent.ID, "offline")
	}()

	log.Printf("agent connected: %s (%s) id=%d", agent.Hostname, ip, agent.ID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var incoming proto.Envelope
		if err := json.Unmarshal(msg, &incoming); err != nil {
			continue
		}
		h.handleAgentMessage(ac, incoming)
	}
}

func (h *AgentWSHandler) handleAgentMessage(ac *hub.AgentConn, env proto.Envelope) {
	switch env.Type {
	case proto.MsgHeartbeat:
		h.store.UpdateAgentLastSeen(ac.AgentID)

	case proto.MsgShellOutput:
		if env.SessionID == "" {
			return
		}
		sess, ok := h.hub.GetSession(env.SessionID)
		if !ok {
			return
		}
		var p proto.ShellOutputPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return
		}
		select {
		case sess.Frontend <- []byte(p.Data):
		default:
		}

	case proto.MsgShellExit:
		if env.SessionID != "" {
			if sess, ok := h.hub.GetSession(env.SessionID); ok {
				select {
				case <-sess.Done:
				default:
					close(sess.Done)
				}
			}
		}

	case proto.MsgFileList, proto.MsgFileChunk, proto.MsgFileAck:
		if env.RequestID != "" {
			h.hub.ResolveRequest(env.RequestID, env)
		}
	}
}
