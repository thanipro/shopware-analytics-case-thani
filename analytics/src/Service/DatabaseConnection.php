<?php

namespace App\Service;

use PDO;

class DatabaseConnection
{
    private ?PDO $pdo = null;

    public function __construct(
        private readonly string $dbPath
    ) {
    }

    public function getConnection(): PDO
    {
        if ($this->pdo === null) {
            $this->pdo = new PDO("sqlite:{$this->dbPath}");
            $this->pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
            $this->pdo->setAttribute(PDO::ATTR_DEFAULT_FETCH_MODE, PDO::FETCH_ASSOC);
            $this->initializeSchema();
        }

        return $this->pdo;
    }

    private function initializeSchema(): void
    {
        $schema = <<<SQL
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
        SQL;

        $this->pdo->exec($schema);
    }
}
