package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?cache=shared&mode=rwc")
	if err != nil {
		return nil, err
	}

	if err := initDB(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
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
	CREATE INDEX IF NOT EXISTS idx_product_id ON events(product_id);
	`

	_, err := db.Exec(schema)
	return err
}
