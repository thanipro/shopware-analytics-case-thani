package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Event struct {
	EventType   string    `json:"event_type" binding:"required"`
	Timestamp   time.Time `json:"timestamp" binding:"required"`
	ProductID   *string   `json:"product_id"`
	OrderAmount *float64  `json:"order_amount"`
}

type Server struct {
	redis *redis.Client
	ctx   context.Context
}

func NewServer() *Server {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return &Server{
		redis: client,
		ctx:   ctx,
	}
}

func (s *Server) HandleEvent(c *gin.Context) {
	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event schema"})
		return
	}

	validTypes := map[string]bool{
		"page_view":    true,
		"add_to_cart":  true,
		"purchase":     true,
	}

	if !validTypes[event.EventType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event type"})
		return
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process event"})
		return
	}

	if err := s.redis.Publish(s.ctx, "analytics:events", eventJSON).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

func (s *Server) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func main() {
	server := NewServer()

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.POST("/events", server.HandleEvent)
		v1.GET("/health", server.HealthCheck)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Ingestion service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
