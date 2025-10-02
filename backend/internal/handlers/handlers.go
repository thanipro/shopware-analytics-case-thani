package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thanipro/shopware-analytics/backend/internal/models"
)

type Handler struct {
	eventQueue chan<- models.Event
}

func New(eventQueue chan<- models.Event) *Handler {
	return &Handler{
		eventQueue: eventQueue,
	}
}

func (h *Handler) HandleEvent(c *gin.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !models.ValidEventTypes[event.EventType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event type"})
		return
	}

	select {
	case h.eventQueue <- event:
		c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Queue full"})
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
