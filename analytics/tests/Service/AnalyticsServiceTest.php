<?php

namespace App\Tests\Service;

use App\Repository\EventRepositoryInterface;
use App\Service\AnalyticsService;
use PHPUnit\Framework\TestCase;

class AnalyticsServiceTest extends TestCase
{
    public function testGetAnalytics(): void
    {
        $repository = $this->createMock(EventRepositoryInterface::class);

        $repository->method('countByType')
            ->willReturnMap([
                ['page_view', 100],
                ['add_to_cart', 30],
                ['purchase', 10],
            ]);

        $repository->method('getPurchaseStats')
            ->willReturn([
                'avg' => 75.50,
                'max' => 299.99,
                'min' => 9.99,
            ]);

        $repository->method('getTopViewedProduct')
            ->willReturn('prod-123');

        $service = new AnalyticsService($repository);
        $analytics = $service->getAnalytics();

        $this->assertEquals(100, $analytics['total_page_views']);
        $this->assertEquals(30, $analytics['total_add_to_carts']);
        $this->assertEquals(10, $analytics['total_purchases']);
        $this->assertEquals(10.0, $analytics['conversion_rate']);
        $this->assertEquals(75.50, $analytics['average_purchase_value']);
        $this->assertEquals(299.99, $analytics['max_purchase_value']);
        $this->assertEquals(9.99, $analytics['min_purchase_value']);
        $this->assertEquals('prod-123', $analytics['top_product_id']);
    }

    public function testGetAnalyticsWithZeroPageViews(): void
    {
        $repository = $this->createMock(EventRepositoryInterface::class);

        $repository->method('countByType')
            ->willReturn(0);

        $repository->method('getPurchaseStats')
            ->willReturn([
                'avg' => 0.0,
                'max' => 0.0,
                'min' => 0.0,
            ]);

        $repository->method('getTopViewedProduct')
            ->willReturn(null);

        $service = new AnalyticsService($repository);
        $analytics = $service->getAnalytics();

        $this->assertEquals(0.0, $analytics['conversion_rate']);
    }
}
