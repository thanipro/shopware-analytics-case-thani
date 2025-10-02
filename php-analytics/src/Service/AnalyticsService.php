<?php

namespace App\Service;

class AnalyticsService
{
    public function __construct(
        private readonly DatabaseConnection $db
    ) {
    }

    public function getAnalytics(): array
    {
        $pdo = $this->db->getPdo();

        $counts = $this->getEventCounts($pdo);
        $purchaseStats = $this->getPurchaseStats($pdo);
        $topProduct = $this->getTopProduct($pdo);

        $totalPageViews = $counts['page_view'] ?? 0;
        $totalPurchases = $counts['purchase'] ?? 0;
        $conversionRate = $totalPageViews > 0
            ? round(($totalPurchases / $totalPageViews) * 100, 2)
            : 0.0;

        return [
            'total_page_views' => $totalPageViews,
            'total_add_to_carts' => $counts['add_to_cart'] ?? 0,
            'total_purchases' => $totalPurchases,
            'conversion_rate' => $conversionRate,
            'average_purchase_value' => $purchaseStats['avg'] ?? 0.0,
            'max_purchase_value' => $purchaseStats['max'] ?? 0.0,
            'min_purchase_value' => $purchaseStats['min'] ?? 0.0,
            'top_product_id' => $topProduct,
        ];
    }

    private function getEventCounts(\PDO $pdo): array
    {
        $stmt = $pdo->query('
            SELECT event_type, COUNT(*) as count
            FROM events
            GROUP BY event_type
        ');

        $counts = [];
        while ($row = $stmt->fetch()) {
            $counts[$row['event_type']] = (int) $row['count'];
        }

        return $counts;
    }

    private function getPurchaseStats(\PDO $pdo): array
    {
        $stmt = $pdo->query('
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

    private function getTopProduct(\PDO $pdo): ?string
    {
        $stmt = $pdo->query('
            SELECT product_id, COUNT(*) as view_count
            FROM events
            WHERE event_type = "page_view" AND product_id IS NOT NULL
            GROUP BY product_id
            ORDER BY view_count DESC
            LIMIT 1
        ');

        $result = $stmt->fetch();
        return $result ? $result['product_id'] : null;
    }
}
