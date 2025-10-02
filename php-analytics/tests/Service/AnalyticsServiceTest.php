<?php

namespace App\Tests\Service;

use App\Service\AnalyticsService;
use App\Service\DatabaseConnection;
use PHPUnit\Framework\TestCase;

class AnalyticsServiceTest extends TestCase
{
    private function createTestDatabase(): DatabaseConnection
    {
        $dbConnection = new DatabaseConnection(':memory:');
        $pdo = $dbConnection->getPdo();

        $pdo->exec('
            CREATE TABLE events (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                event_type TEXT NOT NULL,
                timestamp DATETIME NOT NULL,
                product_id TEXT,
                order_amount REAL,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        ');

        $pdo->exec("
            INSERT INTO events (event_type, timestamp, product_id, order_amount) VALUES
            ('page_view', '2025-10-01 10:00:00', 'prod-1', NULL),
            ('page_view', '2025-10-01 10:01:00', 'prod-2', NULL),
            ('page_view', '2025-10-01 10:02:00', 'prod-1', NULL),
            ('add_to_cart', '2025-10-01 10:03:00', 'prod-1', NULL),
            ('purchase', '2025-10-01 10:04:00', 'prod-1', 99.99),
            ('purchase', '2025-10-01 10:05:00', 'prod-2', 49.99)
        ");

        return $dbConnection;
    }

    public function testGetAnalytics(): void
    {
        $dbConnection = $this->createTestDatabase();
        $service = new AnalyticsService($dbConnection);

        $analytics = $service->getAnalytics();

        $this->assertEquals(3, $analytics['total_page_views']);
        $this->assertEquals(1, $analytics['total_add_to_carts']);
        $this->assertEquals(2, $analytics['total_purchases']);
        $this->assertEquals(66.67, $analytics['conversion_rate']);
        $this->assertEquals(74.99, $analytics['average_purchase_value']);
        $this->assertEquals(99.99, $analytics['max_purchase_value']);
        $this->assertEquals(49.99, $analytics['min_purchase_value']);
        $this->assertEquals('prod-1', $analytics['top_product_id']);
    }
}
