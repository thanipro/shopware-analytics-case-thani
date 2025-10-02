<?php

namespace App\Controller;

use App\Service\AnalyticsService;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\Routing\Attribute\Route;

class AnalyticsController extends AbstractController
{
    public function __construct(
        private readonly AnalyticsService $analyticsService
    ) {
    }

    #[Route('/api/analytics', name: 'analytics', methods: ['GET'])]
    public function getAnalytics(): JsonResponse
    {
        $analytics = $this->analyticsService->getAnalytics();
        return $this->json($analytics);
    }

    #[Route('/api/health', name: 'health', methods: ['GET'])]
    public function health(): JsonResponse
    {
        return $this->json(['status' => 'healthy']);
    }
}
