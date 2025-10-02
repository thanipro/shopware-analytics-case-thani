package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleEvent_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	server := &Server{}
	router.POST("/v1/events", server.HandleEvent)

	invalidJSON := `{"event_type": "page_view", "timestamp": "invalid"`
	req, _ := http.NewRequest("POST", "/v1/events", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleEvent_InvalidEventType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	server := &Server{}
	router.POST("/v1/events", server.HandleEvent)

	event := Event{
		EventType: "invalid_type",
		Timestamp: time.Now(),
	}
	eventJSON, _ := json.Marshal(event)

	req, _ := http.NewRequest("POST", "/v1/events", bytes.NewBuffer(eventJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	server := &Server{}
	router.GET("/v1/health", server.HealthCheck)

	req, _ := http.NewRequest("GET", "/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "healthy", response["status"])
}
