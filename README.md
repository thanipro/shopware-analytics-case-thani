# Shopware Analytics Case Study

A simplified analytics pipeline for e-commerce platforms, demonstrating event ingestion, stream processing, and real-time analytics aggregation.

## Architecture Overview

```
┌──────────────┐
│  E-commerce  │
└──────┬───────┘
       │ POST /v1/events
       ▼
┌─────────────────────────┐
│  Go Ingestion Service   │
│  (Port 8080)            │
│  - Validates events     │
│  - Publishes to Redis   │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│    Redis Queue          │
│  Channel: events        │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  Go Consumer Service    │
│  - Subscribes to Redis  │
│  - Batches events       │
│  - Writes to SQLite     │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│   SQLite Database       │
└──────┬──────────────────┘
       │ Reads
       ▼
┌─────────────────────────┐
│  PHP Analytics Service  │
│  (Port 8000)            │
│  - Computes metrics     │
│  - Serves REST API      │
└──────┬──────────────────┘
       │ GET /api/analytics
       ▼
┌─────────────────────────┐
│  Vue.js Frontend        │
│  (Port 3000)            │
│  - Displays dashboard   │
│  - Auto-refreshes       │
└─────────────────────────┘
```

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Make (optional, for convenience commands)

### Run with Docker Compose

```bash
# Build and start all services
make build && make up

# Or without Make
docker-compose build
docker-compose up -d
```

### Access the Application

- **Frontend Dashboard**: http://localhost:3000
- **Analytics API**: http://localhost:8000/api/analytics
- **Ingestion API**: http://localhost:8080/v1/events

### View Logs

```bash
make logs

# Or
docker-compose logs -f
```

### Stop Services

```bash
make down

# Or
docker-compose down
```

## API Examples

### Submit Events

#### Page View
```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "page_view",
    "timestamp": "2025-10-02T10:30:00Z",
    "product_id": "prod-123"
  }'
```

#### Add to Cart
```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "add_to_cart",
    "timestamp": "2025-10-02T10:35:00Z",
    "product_id": "prod-123"
  }'
```

#### Purchase
```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "purchase",
    "timestamp": "2025-10-02T10:40:00Z",
    "product_id": "prod-123",
    "order_amount": 99.99
  }'
```

### Get Analytics

```bash
curl http://localhost:8000/api/analytics
```

**Response:**
```json
{
  "total_page_views": 150,
  "total_add_to_carts": 45,
  "total_purchases": 12,
  "conversion_rate": 8.0,
  "average_purchase_value": 75.50,
  "max_purchase_value": 299.99,
  "min_purchase_value": 9.99,
  "top_product_id": "prod-123"
}
```

See `api-examples.http` for more examples.

## Project Structure

```
.
├── go-ingestion/          # Go service for event ingestion
│   ├── main.go            # HTTP server + Redis publisher
│   ├── main_test.go       # Unit tests
│   ├── go.mod
│   └── Dockerfile
├── go-consumer/           # Go service for event processing
│   ├── main.go            # Redis consumer + SQLite writer
│   ├── go.mod
│   └── Dockerfile
├── php-analytics/         # Symfony service for analytics
│   ├── src/
│   │   ├── Controller/    # API controllers
│   │   └── Service/       # Business logic
│   ├── tests/             # PHPUnit tests
│   ├── composer.json
│   └── Dockerfile
├── frontend/              # Vue.js dashboard
│   ├── src/
│   │   └── App.vue        # Main component
│   ├── package.json
│   └── Dockerfile
├── docs/                  # Architecture documentation
├── docker-compose.yml     # Multi-service orchestration
├── Makefile              # Convenience commands
└── api-examples.http     # API request examples
```

## Running Tests

### All Tests
```bash
make test
```

### Go Tests Only
```bash
make test-go
# Or
cd go-ingestion && go test -v ./...
```

### PHP Tests Only
```bash
make test-php
# Or
cd php-analytics && composer install && vendor/bin/phpunit
```

## Key Design Decisions

### Why Go for Ingestion?
- **High concurrency**: Handles multiple simultaneous event submissions efficiently
- **Low latency**: Fast response times critical for event tracking
- **Lightweight**: Minimal resource footprint

### Why Redis Queue?
- **Decoupling**: Ingestion service doesn't block on database writes
- **Buffering**: Handles traffic spikes gracefully
- **Reliability**: Events aren't lost if consumer is temporarily down
- **Scalability**: Multiple consumers can process events in parallel

### Why Go for Consumer?
- **Batch processing**: Efficiently batches events for bulk inserts
- **Long-running**: Designed for background processing
- **Shared codebase**: Reuses models from ingestion service

### Why PHP/Symfony for Analytics?
- **Business logic**: Rich ecosystem for complex computations
- **Matches stack**: Aligns with Shopware's technology choices
- **Separation of concerns**: Analytics computation separate from event processing

### Why SQLite?
- **Simplicity**: No external database server needed for demo
- **Local development**: Easy to run and inspect
- **Production path**: Same code works with PostgreSQL/MySQL by changing connection string

### Event Schema

```json
{
  "event_type": "page_view|add_to_cart|purchase",
  "timestamp": "2025-10-02T10:30:00Z",
  "product_id": "prod-123",        // Optional
  "order_amount": 99.99            // Required for purchases
}
```

## Assumptions & Trade-offs

### Assumptions
1. **Events are idempotent**: No deduplication logic (for simplicity)
2. **Single tenant**: No multi-shop separation
3. **Conversion rate**: Calculated as purchases/page_views (simplified funnel)
4. **Top product**: Based on page views, not purchases
5. **Time zone**: All timestamps assumed UTC

### Trade-offs
1. **SQLite vs PostgreSQL**
   - ✅ Simple setup, no external database
   - ❌ Limited concurrent writes
   - **Production**: Would use PostgreSQL/MySQL

2. **Eventual consistency**
   - ✅ Fast ingestion response times
   - ❌ Dashboard may lag 1-5 seconds
   - **Acceptable**: Analytics don't need real-time guarantees

3. **Batch processing**
   - ✅ Efficient database writes (100 events or 5 seconds)
   - ❌ Small delay before events appear in analytics
   - **Tunable**: Batch size and interval configurable

4. **In-memory batching**
   - ✅ Fast aggregation
   - ❌ Events lost if consumer crashes before flush
   - **Production**: Would use persistent queue (Kafka/SQS)

5. **Shared database file**
   - ✅ Simple for local development
   - ❌ Coupling between consumer and analytics
   - **Production**: Consumer would expose Data Access API

## Local Development (Without Docker)

### 1. Start Redis
```bash
redis-server
```

### 2. Start Go Ingestion
```bash
cd go-ingestion
go run main.go
```

### 3. Start Go Consumer
```bash
cd go-consumer
go run main.go
```

### 4. Start PHP Analytics
```bash
cd php-analytics
composer install
php -S localhost:8000 -t public
```

### 5. Start Frontend
```bash
cd frontend
npm install
npm run dev
```

## Monitoring & Observability

Currently implemented:
- Health check endpoints (`/v1/health`, `/api/health`)
- Structured logging in all services
- Request/response logging

Production additions (see `/docs/observability.md`):
- Prometheus metrics
- Distributed tracing (OpenTelemetry)
- Centralized logging (ELK stack)
- Alerting (PagerDuty/Slack)

## Scalability Considerations

See `/docs/scalability.md` for detailed discussion.

**Summary**:
- Replace Redis Pub/Sub with Kafka/Kinesis for persistence
- Horizontal scaling of ingestion and consumer services
- Migrate to time-series database (ClickHouse/TimescaleDB)
- Implement CQRS pattern for read/write separation
- Add caching layer (Redis) for frequently accessed metrics

## Security Considerations

See `/docs/security.md` for detailed discussion.

Current state: **No authentication** (demo purposes)

Production requirements:
- API key authentication for event submission
- Rate limiting per client
- TLS/HTTPS encryption
- Input validation and sanitization
- PII anonymization (IP addresses, user IDs)
- Secrets management (AWS Secrets Manager)

## Future Improvements

1. **User tracking**: Session-based conversion funnels
2. **Time-based analytics**: Hourly/daily/weekly aggregations
3. **Real-time dashboards**: WebSocket updates instead of polling
4. **Advanced metrics**: Bounce rate, session duration, funnel analysis
5. **Data retention**: Automated archival of old events
6. **Multi-tenancy**: Shop-level data isolation
7. **Export functionality**: CSV/PDF report generation
8. **Admin interface**: Event inspection, replay, filtering

## License

This is a case study project for Shopware/CtrlAltElite.

---

**Built with**: Go (Gin), PHP (Symfony), Vue.js, Redis, SQLite
**Author**: Oluwaseun Thani
**Date**: October 2025
