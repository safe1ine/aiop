package connector

import (
	"encoding/json"
	"log"
	"math"
	"net"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"github.com/aipo/agent/internal/config"
	"github.com/aipo/agent/internal/executor"
	"github.com/aipo/agent/internal/filemgr"
)

type MessageType = string

const (
	MsgRegister    MessageType = "register"
	MsgHeartbeat   MessageType = "heartbeat"
	MsgShellOutput MessageType = "shell_output"
	MsgShellExit   MessageType = "shell_exit"
	MsgShellStart  MessageType = "shell_start"
	MsgShellInput  MessageType = "shell_input"
	MsgShellResize MessageType = "shell_resize"
	MsgShellStop   MessageType = "shell_stop"
)

type Envelope struct {
	Type      string          `json:"type"`
	SessionID string          `json:"session_id,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

type Client struct {
	cfg    *config.Config
	shells map[string]*executor.Shell
	conn   *websocket.Conn
	sendCh chan []byte // serialized writes
}

func New(cfg *config.Config) *Client {
	return &Client{cfg: cfg, shells: make(map[string]*executor.Shell), sendCh: make(chan []byte, 256)}
}

func (c *Client) Run() {
	backoff := time.Second
	for {
		if err := c.connect(); err != nil {
			log.Printf("connection error: %v, retrying in %s", err, backoff)
		}
		time.Sleep(backoff)
		backoff = time.Duration(math.Min(float64(backoff*2), float64(60*time.Second)))
	}
}

func (c *Client) connect() error {
	u, err := url.Parse(c.cfg.ServerURL)
	if err != nil {
		return err
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	c.conn = conn

	// Single writer goroutine — websocket.Conn is not concurrent-safe for writes
	stopWriter := make(chan struct{})
	go func() {
		for {
			select {
			case data := <-c.sendCh:
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					return
				}
			case <-stopWriter:
				return
			}
		}
	}()
	defer close(stopWriter)

	hostname, _ := os.Hostname()
	ip := getOutboundIP()

	regPayload, _ := json.Marshal(map[string]string{
		"enroll_token": c.cfg.EnrollToken,
		"hostname":     hostname,
		"ip":           ip,
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
	})
	if err := c.send(Envelope{Type: MsgRegister, Payload: regPayload}); err != nil {
		return err
	}
	log.Printf("registered: hostname=%s ip=%s", hostname, ip)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	msgCh := make(chan Envelope, 32)
	errCh := make(chan error, 1)
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			var env Envelope
			if err := json.Unmarshal(data, &env); err != nil {
				continue
			}
			msgCh <- env
		}
	}()

	for {
		select {
		case err := <-errCh:
			return err
		case <-ticker.C:
			hbPayload, _ := json.Marshal(map[string]int64{"timestamp": time.Now().Unix()})
			c.send(Envelope{Type: MsgHeartbeat, Payload: hbPayload})
		case env := <-msgCh:
			c.handle(env)
		}
	}
}

func (c *Client) handle(env Envelope) {
	switch env.Type {
	case MsgShellStart:
		var p struct {
			Cols uint16 `json:"cols"`
			Rows uint16 `json:"rows"`
		}
		json.Unmarshal(env.Payload, &p)
		if p.Cols == 0 {
			p.Cols = 80
		}
		if p.Rows == 0 {
			p.Rows = 24
		}
		sh, err := executor.NewShell(p.Cols, p.Rows)
		if err != nil {
			log.Printf("shell start error: %v", err)
			return
		}
		c.shells[env.SessionID] = sh
		go c.streamShell(env.SessionID, sh)

	case MsgShellInput:
		sh, ok := c.shells[env.SessionID]
		if !ok {
			return
		}
		var p struct {
			Data string `json:"data"`
		}
		json.Unmarshal(env.Payload, &p)
		sh.Write([]byte(p.Data))

	case MsgShellResize:
		sh, ok := c.shells[env.SessionID]
		if !ok {
			return
		}
		var p struct {
			Cols uint16 `json:"cols"`
			Rows uint16 `json:"rows"`
		}
		json.Unmarshal(env.Payload, &p)
		sh.Resize(p.Cols, p.Rows)

	case MsgShellStop:
		if sh, ok := c.shells[env.SessionID]; ok {
			sh.Close()
			delete(c.shells, env.SessionID)
		}

	case "file_list_req":
		go c.handleFileList(env)
	case "file_download_req":
		go c.handleFileDownload(env)
	case "file_upload":
		go c.handleFileUpload(env)
	case "file_delete":
		go c.handleFileDelete(env)
	}
}

func (c *Client) streamShell(sessionID string, sh *executor.Shell) {
	buf := make([]byte, 4096)
	for {
		n, err := sh.Read(buf)
		if n > 0 {
			payload, _ := json.Marshal(map[string]string{"data": string(buf[:n])})
			c.send(Envelope{
				Type:      MsgShellOutput,
				SessionID: sessionID,
				Payload:   payload,
			})
		}
		if err != nil {
			exitPayload, _ := json.Marshal(map[string]int{"code": 0})
			c.send(Envelope{
				Type:      MsgShellExit,
				SessionID: sessionID,
				Payload:   exitPayload,
			})
			delete(c.shells, sessionID)
			return
		}
	}
}

func (c *Client) send(env Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	select {
	case c.sendCh <- data:
	default:
		// channel full, drop (shouldn't happen in practice)
	}
	return nil
}

// getOutboundIP returns the preferred outbound IP of this machine
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func (c *Client) sendAck(requestID string, errMsg string) {
	type ack struct {
		Error string `json:"error,omitempty"`
	}
	payload, _ := json.Marshal(ack{Error: errMsg})
	c.send(Envelope{Type: "file_ack", RequestID: requestID, Payload: payload})
}

func (c *Client) handleFileList(env Envelope) {
	var p struct {
		Path string `json:"path"`
	}
	json.Unmarshal(env.Payload, &p)
	if p.Path == "" {
		p.Path = "/"
	}
	entries, err := filemgr.ListDir(p.Path)
	if err != nil {
		entries = []filemgr.FileEntry{}
	}
	resp, _ := filemgr.MarshalListResponse(p.Path, entries)
	c.send(Envelope{Type: "file_list", RequestID: env.RequestID, Payload: resp})
}

func (c *Client) handleFileDownload(env Envelope) {
	var p struct {
		Path string `json:"path"`
	}
	json.Unmarshal(env.Payload, &p)

	const chunkSize = 512 * 1024 // 512KB
	chunks, err := filemgr.ReadFileChunked(p.Path, chunkSize)
	if err != nil {
		c.sendAck(env.RequestID, err.Error())
		return
	}
	total := len(chunks)
	for i, chunk := range chunks {
		type chunkMsg struct {
			Path  string `json:"path"`
			Index int    `json:"index"`
			Total int    `json:"total"`
			Data  []byte `json:"data"`
		}
		payload, _ := json.Marshal(chunkMsg{Path: p.Path, Index: i, Total: total, Data: chunk})
		c.send(Envelope{Type: "file_chunk", RequestID: env.RequestID, Payload: payload})
	}
}

func (c *Client) handleFileUpload(env Envelope) {
	var p struct {
		Path string `json:"path"`
		Data []byte `json:"data"`
	}
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		c.sendAck(env.RequestID, err.Error())
		return
	}
	if err := filemgr.WriteFile(p.Path, p.Data); err != nil {
		c.sendAck(env.RequestID, err.Error())
		return
	}
	c.sendAck(env.RequestID, "")
}

func (c *Client) handleFileDelete(env Envelope) {
	var p struct {
		Path string `json:"path"`
	}
	json.Unmarshal(env.Payload, &p)
	err := filemgr.DeletePath(p.Path)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	c.sendAck(env.RequestID, errMsg)
}
