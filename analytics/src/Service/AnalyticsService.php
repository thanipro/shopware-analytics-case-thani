<?php

namespace App\Service;

use App\Repository\EventRepositoryInterface;

class AnalyticsService
{
    public function __construct(
        private readonly EventRepositoryInterface $eventRepository
    ) {
    }

    public function getAnalytics(): array
    {
        $pageViews = $this->eventRepository->countByType('page_view');
        $addToCarts = $this->eventRepository->countByType('add_to_cart');
        $purchases = $this->eventRepository->countByType('purchase');

        $purchaseStats = $this->eventRepository->getPurchaseStats();
        $topProduct = $this->eventRepository->getTopViewedProduct();

        $conversionRate = $pageViews > 0
            ? (float) round(($purchases / $pageViews) * 100, 2)
            : 0.0;

        return [
            'total_page_views' => $pageViews,
            'total_add_to_carts' => $addToCarts,
            'total_purchases' => $purchases,
            'conversion_rate' => $conversionRate,
            'average_purchase_value' => $purchaseStats['avg'],
            'max_purchase_value' => $purchaseStats['max'],
            'min_purchase_value' => $purchaseStats['min'],
            'top_product_id' => $topProduct,
        ];
    }
}
