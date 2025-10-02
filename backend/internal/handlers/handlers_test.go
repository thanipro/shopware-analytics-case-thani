package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/thanipro/shopware-analytics/backend/internal/models"
)

func TestHandleEvent_ValidEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queue := make(chan models.Event, 10)
	handler := New(queue)

	router := gin.New()
	router.POST("/events", handler.HandleEvent)

	event := map[string]interface{}{
		"event_type": "page_view",
		"timestamp":  "2025-10-02T10:30:00Z",
		"product_id": "prod-123",
	}

	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	select {
	case <-queue:
		// Event received
	default:
		t.Error("Event was not queued")
	}
}

func TestHandleEvent_InvalidEventType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queue := make(chan models.Event, 10)
	handler := New(queue)

	router := gin.New()
	router.POST("/events", handler.HandleEvent)

	event := map[string]interface{}{
		"event_type": "invalid_type",
		"timestamp":  "2025-10-02T10:30:00Z",
	}

	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queue := make(chan models.Event, 10)
	handler := New(queue)

	router := gin.New()
	router.GET("/health", handler.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
}
