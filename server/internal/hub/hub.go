package hub

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/aipo/server/internal/proto"
)

// AgentConn represents a connected agent
type AgentConn struct {
	AgentID  int64
	Name     string
	Conn     *websocket.Conn
	mu       sync.Mutex
	sessions map[string]*TerminalSession // sessionID -> session
}

func (a *AgentConn) Send(env proto.Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Conn.WriteMessage(websocket.TextMessage, data)
}

// TerminalSession bridges frontend WS <-> agent WS
type TerminalSession struct {
	ID       string
	AgentID  int64
	Frontend chan []byte // data to send to frontend
	Done     chan struct{}
}

// Hub manages all agent connections
type Hub struct {
	mu       sync.RWMutex
	agents   map[int64]*AgentConn
	sessions map[string]*TerminalSession
	pending  map[string]chan proto.Envelope // requestID -> response channel
}

func New() *Hub {
	return &Hub{
		agents:   make(map[int64]*AgentConn),
		sessions: make(map[string]*TerminalSession),
		pending:  make(map[string]chan proto.Envelope),
	}
}

func (h *Hub) Register(agentID int64, name string, conn *websocket.Conn) *AgentConn {
	ac := &AgentConn{
		AgentID:  agentID,
		Name:     name,
		Conn:     conn,
		sessions: make(map[string]*TerminalSession),
	}
	h.mu.Lock()
	h.agents[agentID] = ac
	h.mu.Unlock()
	return ac
}

func (h *Hub) Unregister(agentID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ac, ok := h.agents[agentID]; ok {
		// close all sessions for this agent
		for _, sess := range ac.sessions {
			close(sess.Done)
			delete(h.sessions, sess.ID)
		}
		delete(h.agents, agentID)
	}
}

func (h *Hub) GetAgent(agentID int64) (*AgentConn, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ac, ok := h.agents[agentID]
	return ac, ok
}

func (h *Hub) IsOnline(agentID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.agents[agentID]
	return ok
}

func (h *Hub) AddSession(sess *TerminalSession) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[sess.ID] = sess
	if ac, ok := h.agents[sess.AgentID]; ok {
		ac.sessions[sess.ID] = sess
	}
}

func (h *Hub) GetSession(sessionID string) (*TerminalSession, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	sess, ok := h.sessions[sessionID]
	return sess, ok
}

func (h *Hub) RemoveSession(sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if sess, ok := h.sessions[sessionID]; ok {
		if ac, ok := h.agents[sess.AgentID]; ok {
			delete(ac.sessions, sessionID)
		}
		delete(h.sessions, sessionID)
	}
}

// RegisterRequest creates a buffered channel for a pending request and returns it.
func (h *Hub) RegisterRequest(requestID string) chan proto.Envelope {
	ch := make(chan proto.Envelope, 512)
	h.mu.Lock()
	h.pending[requestID] = ch
	h.mu.Unlock()
	return ch
}

// ResolveRequest delivers an envelope to the waiting HTTP handler.
func (h *Hub) ResolveRequest(requestID string, env proto.Envelope) {
	h.mu.RLock()
	ch, ok := h.pending[requestID]
	h.mu.RUnlock()
	if ok {
		select {
		case ch <- env:
		default:
		}
	}
}

// CancelRequest closes the channel so the HTTP handler unblocks.
func (h *Hub) CancelRequest(requestID string) {
	h.mu.Lock()
	if ch, ok := h.pending[requestID]; ok {
		close(ch)
		delete(h.pending, requestID)
	}
	h.mu.Unlock()
}

// FinishRequest removes the request from the pending map (called by HTTP handler when done).
func (h *Hub) FinishRequest(requestID string) {
	h.mu.Lock()
	delete(h.pending, requestID)
	h.mu.Unlock()
}
