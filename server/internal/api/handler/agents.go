package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/aipo/server/internal/db"
	"github.com/aipo/server/internal/hub"
)

type AgentHandler struct {
	store       db.Store
	hub         *hub.Hub
	enrollToken string
}

func NewAgentHandler(store db.Store, h *hub.Hub, enrollToken string) *AgentHandler {
	return &AgentHandler{store: store, hub: h, enrollToken: enrollToken}
}

func (h *AgentHandler) List(c *gin.Context) {
	agents, err := h.store.ListAgents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, a := range agents {
		if h.hub.IsOnline(a.ID) {
			a.Status = "online"
		} else {
			a.Status = "offline"
		}
	}
	c.JSON(http.StatusOK, agents)
}

func (h *AgentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.store.DeleteAgent(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// EnrollToken returns the server's enroll token for generating install commands
func (h *AgentHandler) EnrollToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enroll_token": h.enrollToken})
}
