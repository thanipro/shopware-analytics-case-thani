# Quick Start Guide

## Prerequisites

- Docker & Docker Compose
- OR: Go 1.21+, PHP 8.2+, Node.js 20+, Redis

## Option 1: Docker Compose (Recommended)

```bash
# Build all services
docker-compose build

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Option 2: Local Development

### Terminal 1 - Redis
```bash
redis-server
```

### Terminal 2 - Go Ingestion Service
```bash
cd go-ingestion
go mod download
go run main.go
```

### Terminal 3 - Go Consumer Service
```bash
cd go-consumer
go mod download
mkdir -p ../data
DB_PATH=../data/analytics.db go run main.go
```

### Terminal 4 - PHP Analytics Service
```bash
cd php-analytics
composer install
DB_PATH=../data/analytics.db php -S localhost:8000 -t public
```

### Terminal 5 - Frontend
```bash
cd frontend
npm install
VITE_API_URL=http://localhost:8000 npm run dev
```

## Testing the Application

### 1. Submit Test Events

```bash
# Page view
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "page_view",
    "timestamp": "2025-10-02T10:00:00Z",
    "product_id": "prod-123"
  }'

# Add to cart
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "add_to_cart",
    "timestamp": "2025-10-02T10:05:00Z",
    "product_id": "prod-123"
  }'

# Purchase
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "purchase",
    "timestamp": "2025-10-02T10:10:00Z",
    "product_id": "prod-123",
    "order_amount": 99.99
  }'
```

### 2. Check Analytics

```bash
curl http://localhost:8000/api/analytics | jq
```

### 3. View Dashboard

Open http://localhost:3000 in your browser

## Running Tests

```bash
# All tests
make test

# Go tests only
cd go-ingestion && go test -v ./...

# PHP tests only
cd php-analytics && composer install && vendor/bin/phpunit
```

## Troubleshooting

### "Connection refused" errors
- Ensure Redis is running: `redis-cli ping`
- Check if ports are available: `lsof -i :8080,8000,3000,6379`

### Database errors
- Ensure data directory exists: `mkdir -p data`
- Check permissions: `chmod 755 data`

### Go module errors
- Run: `go mod tidy` in go-ingestion and go-consumer directories

### PHP dependencies
- Run: `composer install` in php-analytics directory

### Node modules
- Run: `npm install` in frontend directory

## Architecture

```
E-commerce → Ingestion (8080) → Redis → Consumer → SQLite
                                                       ↓
Frontend (3000) ← Analytics (8000) ←──────────────────┘
```
