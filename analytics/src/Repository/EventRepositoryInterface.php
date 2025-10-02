<?php

namespace App\Repository;

interface EventRepositoryInterface
{
    public function countByType(string $type): int;

    public function getPurchaseStats(): array;

    public function getTopViewedProduct(): ?string;
}
