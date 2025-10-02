<?php

namespace App\Repository;

use PDO;

class EventRepository implements EventRepositoryInterface
{
    public function __construct(
        private readonly PDO $pdo
    ) {
    }

    public function countByType(string $type): int
    {
        $stmt = $this->pdo->prepare(
            'SELECT COUNT(*) as count FROM events WHERE event_type = :type'
        );

        $stmt->execute(['type' => $type]);
        $result = $stmt->fetch();

        return (int) ($result['count'] ?? 0);
    }

    public function getPurchaseStats(): array
    {
        $stmt = $this->pdo->query('
            SELECT
                AVG(order_amount) as avg,
                MAX(order_amount) as max,
                MIN(order_amount) as min
            FROM events
            WHERE event_type = "purchase" AND order_amount IS NOT NULL
        ');

        $stats = $stmt->fetch();

        return [
            'avg' => $stats['avg'] ? round((float) $stats['avg'], 2) : 0.0,
            'max' => $stats['max'] ? round((float) $stats['max'], 2) : 0.0,
            'min' => $stats['min'] ? round((float) $stats['min'], 2) : 0.0,
        ];
    }

    public function getTopViewedProduct(): ?string
    {
        $stmt = $this->pdo->query('
            SELECT product_id
            FROM events
            WHERE event_type = "page_view" AND product_id IS NOT NULL
            GROUP BY product_id
            ORDER BY COUNT(*) DESC
            LIMIT 1
        ');

        $result = $stmt->fetch();
        return $result ? $result['product_id'] : null;
    }
}
