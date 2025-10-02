package consumer

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/thanipro/shopware-analytics/backend/internal/models"
)

type Consumer struct {
	db         *sql.DB
	eventQueue chan models.Event
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func New(db *sql.DB, queueSize int) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		db:         db,
		eventQueue: make(chan models.Event, queueSize),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (c *Consumer) Queue() chan<- models.Event {
	return c.eventQueue
}

func (c *Consumer) Start() {
	batch := make([]models.Event, 0, 100)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Event consumer started")

	for {
		select {
		case event := <-c.eventQueue:
			batch = append(batch, event)

			if len(batch) >= 100 {
				c.flushBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				c.flushBatch(batch)
				batch = batch[:0]
			}

		case <-c.ctx.Done():
			if len(batch) > 0 {
				c.flushBatch(batch)
			}
			return
		}
	}
}

func (c *Consumer) flushBatch(batch []models.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO events (event_type, timestamp, product_id, order_amount)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		log.Printf("Failed to prepare statement: %v", err)
		return
	}
	defer stmt.Close()

	for _, event := range batch {
		_, err := stmt.Exec(event.EventType, event.Timestamp, event.ProductID, event.OrderAmount)
		if err != nil {
			log.Printf("Failed to insert event: %v", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	log.Printf("Flushed %d events to database", len(batch))
}

func (c *Consumer) Shutdown() {
	c.cancel()
}
