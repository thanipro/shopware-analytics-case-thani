# Architecture Decisions

Quick notes on why I built things the way I did.

## Why Go for the core event receiving part?

I needed something fast that could handle lots of concurrent requests. Go's goroutines made this easy, and the built-in channels gave me a simple in-memory queue without adding Redis. 


## Why Symfony for Analytics?

The requirement mentioned matching Shopware's stack, and I wanted to show understanding os symfony. I went with it because:
- The repository pattern keeps database queries separate from business logic
- Makes testing easier (mock the repository)
- Dependency injection is built-in

The analytics service just needs to read data and calculate metrics - PHP is fine for this.

**Alternative considered:** Building it in Go too - but keeping PHP shows I can work with different stacks, and the repository pattern demonstration seemed valuable.

## Internal Queue vs Redis

Initially thought about Redis pub/sub, but the requirement said "no complex pipeline." Go channels are:
- Already in-memory (fast)
- No extra infrastructure needed
- Good enough for the demo scale

If this needed to handle millions of events, I'd switch to Kafka for persistence and replay capability.

## Batch Processing

Writing events one-by-one would kill the database. So I batch them:
- **100 events** OR **5 seconds** (whichever comes first)
- Uses a single transaction for all inserts
- Means there's a small delay, but that's fine for analytics

The 5-second ticker ensures events don't sit in the queue too long during low traffic.

## Repository Pattern in PHP

```php
EventRepositoryInterface → EventRepository → AnalyticsService
```

This keeps things testable. The service doesn't care if data comes from SQLite, PostgreSQL, or a mock - it just asks the repository. Made writing unit tests much easier.

## SQLite for Storage

Simple choice for a demo:
- No setup needed
- File-based
- Same SQL works with PostgreSQL later

The downside is concurrent writes are limited, so this wouldn't work at real scale. For production, I'd use PostgreSQL with read replicas.

## TypeScript for Frontend

Added TypeScript because:
- Type safety catches bugs early
- The `Analytics` interface makes it clear what the API returns
- It's becoming standard for Vue projects

Could've used plain JavaScript, but TypeScript doesn't add much complexity and the types help.

## CORS Handling

Made a Symfony EventSubscriber to add CORS headers globally instead of in every controller. Cleaner and means I don't forget to add them later.

## What I'd Change for Real Traffic

**Right now:** Good for up to ~10K events/second

**For production:**
1. PostgreSQL instead of SQLite
2. Kafka instead of in-memory channels (persistence + replay)
3. Multiple backend instances behind a load balancer
4. Read replicas for the analytics queries
5. Redis cache for the computed metrics (5-10 second TTL)
6. API authentication (API keys or OAuth)
7. Proper monitoring (Prometheus + Grafana)

## Testing Strategy

- **Go:** Unit tests for handlers (mock the queue)
- **PHP:** Unit tests with mocked repositories, integration tests with real SQLite
- **E2E:** Manually tested the full flow (would add Playwright for real)

Focused on testing the business logic more than infrastructure code.

## Things I Kept Simple On Purpose

- No authentication (demo only)
- No rate limiting
- No retry logic for failed batches
- Single database file shared between services
- No data retention policies

These would all be needed for production, but they'd distract from demonstrating the core architecture.
