package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/aipo/server/internal/hub"
	"github.com/aipo/server/internal/proto"
)

type FileHandler struct {
	hub *hub.Hub
}

func NewFileHandler(h *hub.Hub) *FileHandler {
	return &FileHandler{hub: h}
}

func (h *FileHandler) getAgent(c *gin.Context) (*hub.AgentConn, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent id"})
		return nil, false
	}
	ac, ok := h.hub.GetAgent(id)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent offline"})
		return nil, false
	}
	return ac, true
}

// List handles GET /api/v1/agents/:id/files?path=/some/dir
func (h *FileHandler) List(c *gin.Context) {
	ac, ok := h.getAgent(c)
	if !ok {
		return
	}
	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	reqID := uuid.New().String()
	ch := h.hub.RegisterRequest(reqID)
	defer h.hub.FinishRequest(reqID)

	payload, _ := json.Marshal(proto.FileRequestPayload{Path: path})
	if err := ac.Send(proto.Envelope{Type: proto.MsgFileListReq, RequestID: reqID, Payload: payload}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "send failed"})
		return
	}

	select {
	case env, ok := <-ch:
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent disconnected"})
			return
		}
		var p proto.FileListPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "bad response"})
			return
		}
		c.JSON(http.StatusOK, p)
	case <-time.After(15 * time.Second):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout"})
	}
}

// Download handles GET /api/v1/agents/:id/files/download?path=/some/file
func (h *FileHandler) Download(c *gin.Context) {
	ac, ok := h.getAgent(c)
	if !ok {
		return
	}
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}

	reqID := uuid.New().String()
	ch := h.hub.RegisterRequest(reqID)
	defer h.hub.FinishRequest(reqID)

	payload, _ := json.Marshal(proto.FileRequestPayload{Path: path})
	if err := ac.Send(proto.Envelope{Type: proto.MsgFileDownloadReq, RequestID: reqID, Payload: payload}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "send failed"})
		return
	}

	var buf bytes.Buffer
	timeout := time.After(60 * time.Second)
	for {
		select {
		case env, ok := <-ch:
			if !ok {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent disconnected"})
				return
			}
			var chunk proto.FileChunkPayload
			if err := json.Unmarshal(env.Payload, &chunk); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "bad chunk"})
				return
			}
			buf.Write(chunk.Data)
			if chunk.Index == chunk.Total-1 {
				c.Header("Content-Disposition", `attachment; filename="`+filepath.Base(path)+`"`)
				c.Data(http.StatusOK, "application/octet-stream", buf.Bytes())
				return
			}
		case <-timeout:
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout"})
			return
		}
	}
}

// Upload handles POST /api/v1/agents/:id/files/upload?path=/dest/dir
func (h *FileHandler) Upload(c *gin.Context) {
	ac, ok := h.getAgent(c)
	if !ok {
		return
	}
	destDir := c.Query("path")
	if destDir == "" {
		destDir = "/"
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read failed"})
		return
	}

	destPath := filepath.Join(destDir, header.Filename)
	reqID := uuid.New().String()
	ch := h.hub.RegisterRequest(reqID)
	defer h.hub.FinishRequest(reqID)

	payload, _ := json.Marshal(proto.FileUploadPayload{Path: destPath, Data: buf.Bytes()})
	if err := ac.Send(proto.Envelope{Type: proto.MsgFileUpload, RequestID: reqID, Payload: payload}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "send failed"})
		return
	}

	select {
	case env, ok := <-ch:
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent disconnected"})
			return
		}
		var ack proto.FileAckPayload
		json.Unmarshal(env.Payload, &ack)
		if ack.Error != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ack.Error})
			return
		}
		c.JSON(http.StatusOK, gin.H{"path": destPath})
	case <-time.After(60 * time.Second):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout"})
	}
}

// Delete handles DELETE /api/v1/agents/:id/files?path=/some/file
func (h *FileHandler) Delete(c *gin.Context) {
	ac, ok := h.getAgent(c)
	if !ok {
		return
	}
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}

	reqID := uuid.New().String()
	ch := h.hub.RegisterRequest(reqID)
	defer h.hub.FinishRequest(reqID)

	payload, _ := json.Marshal(proto.FileDeletePayload{Path: path})
	if err := ac.Send(proto.Envelope{Type: proto.MsgFileDelete, RequestID: reqID, Payload: payload}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "send failed"})
		return
	}

	select {
	case env, ok := <-ch:
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent disconnected"})
			return
		}
		var ack proto.FileAckPayload
		json.Unmarshal(env.Payload, &ack)
		if ack.Error != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ack.Error})
			return
		}
		c.Status(http.StatusNoContent)
	case <-time.After(15 * time.Second):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout"})
	}
}
