package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/aipo/server/internal/hub"
	"github.com/aipo/server/internal/proto"
)

var termUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type TerminalHandler struct {
	hub *hub.Hub
}

func NewTerminalHandler(h *hub.Hub) *TerminalHandler {
	return &TerminalHandler{hub: h}
}

func (h *TerminalHandler) Connect(c *gin.Context) {
	agentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent id"})
		return
	}

	ac, ok := h.hub.GetAgent(agentID)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent offline"})
		return
	}

	conn, err := termUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	sessionID := uuid.New().String()
	sess := &hub.TerminalSession{
		ID:       sessionID,
		AgentID:  agentID,
		Frontend: make(chan []byte, 256),
		Done:     make(chan struct{}),
	}
	h.hub.AddSession(sess)
	defer h.hub.RemoveSession(sessionID)

	// Tell agent to start shell
	startPayload, _ := json.Marshal(proto.ShellStartPayload{Cols: 80, Rows: 24})
	ac.Send(proto.Envelope{
		Type:      proto.MsgShellStart,
		SessionID: sessionID,
		Payload:   startPayload,
	})

	// Frontend -> Agent
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				select {
				case <-sess.Done:
				default:
					close(sess.Done)
				}
				return
			}
			// Resize: JSON with cols/rows fields
			var resize proto.ShellResizePayload
			if json.Unmarshal(data, &resize) == nil && resize.Cols > 0 && resize.Rows > 0 {
				p, _ := json.Marshal(resize)
				ac.Send(proto.Envelope{
					Type:      proto.MsgShellResize,
					SessionID: sessionID,
					Payload:   p,
				})
				continue
			}
			// Raw input
			p, _ := json.Marshal(proto.ShellInputPayload{Data: string(data)})
			ac.Send(proto.Envelope{
				Type:      proto.MsgShellInput,
				SessionID: sessionID,
				Payload:   p,
			})
		}
	}()

	// Agent -> Frontend
	for {
		select {
		case <-sess.Done:
			return
		case data := <-sess.Frontend:
			if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				return
			}
		}
	}
}
