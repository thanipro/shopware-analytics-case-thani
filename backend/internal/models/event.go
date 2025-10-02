package models

import "time"

type Event struct {
	EventType   string    `json:"event_type" binding:"required"`
	Timestamp   time.Time `json:"timestamp" binding:"required"`
	ProductID   *string   `json:"product_id"`
	OrderAmount *float64  `json:"order_amount"`
}

var ValidEventTypes = map[string]bool{
	"page_view":   true,
	"add_to_cart": true,
	"purchase":    true,
}
