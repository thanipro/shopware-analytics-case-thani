package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	_ "github.com/mattn/go-sqlite3"
)

type Event struct {
	EventType   string    `json:"event_type"`
	Timestamp   time.Time `json:"timestamp"`
	ProductID   *string   `json:"product_id"`
	OrderAmount *float64  `json:"order_amount"`
}

type Consumer struct {
	redis *redis.Client
	db    *sql.DB
	ctx   context.Context
	batch []Event
	mu    sync.Mutex
}

func NewConsumer() *Consumer {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/analytics.db"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := initDB(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	return &Consumer{
		redis: redisClient,
		db:    db,
		ctx:   ctx,
		batch: make([]Event, 0, 100),
	}
}

func initDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		product_id TEXT,
		order_amount REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_event_type ON events(event_type);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_product_id ON events(product_id);
	`

	_, err := db.Exec(schema)
	return err
}

func (c *Consumer) Start() {
	pubsub := c.redis.Subscribe(c.ctx, "analytics:events")
	defer pubsub.Close()

	channel := pubsub.Channel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Consumer started, waiting for events...")

	for {
		select {
		case msg := <-channel:
			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("Failed to unmarshal event: %v", err)
				continue
			}

			c.mu.Lock()
			c.batch = append(c.batch, event)
			shouldFlush := len(c.batch) >= 100
			c.mu.Unlock()

			if shouldFlush {
				c.flush()
			}

		case <-ticker.C:
			c.flush()

		case <-sigChan:
			log.Println("Shutting down consumer...")
			c.flush()
			c.db.Close()
			c.redis.Close()
			return
		}
	}
}

func (c *Consumer) flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.batch) == 0 {
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	stmt, err := tx.Prepare(`
		INSERT INTO events (event_type, timestamp, product_id, order_amount)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to prepare statement: %v", err)
		return
	}
	defer stmt.Close()

	for _, event := range c.batch {
		_, err := stmt.Exec(event.EventType, event.Timestamp, event.ProductID, event.OrderAmount)
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to insert event: %v", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	log.Printf("Flushed %d events to database", len(c.batch))
	c.batch = c.batch[:0]
}

func main() {
	consumer := NewConsumer()
	consumer.Start()
}
