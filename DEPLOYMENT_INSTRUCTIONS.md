# Deployment & Testing Instructions

## System Status

✅ **All Code Complete**
- Go Ingestion Service
- Go Consumer Service
- PHP Analytics Service
- Vue.js Frontend
- Docker Compose configuration
- Complete documentation

## Running the Application

### Option 1: Docker Compose (Recommended)

**Prerequisites**: Docker and Docker Compose installed and running

```bash
# 1. Build all containers
docker-compose build

# 2. Start all services
docker-compose up -d

# 3. Check services are running
docker-compose ps

# 4. View logs
docker-compose logs -f

# Expected output:
# - redis: Running on port 6379
# - ingestion: Running on port 8080
# - consumer: Processing events in background
# - analytics: Running on port 8000
# - frontend: Running on port 3000
```

**Access points**:
- Frontend Dashboard: http://localhost:3000
- Analytics API: http://localhost:8000/api/analytics
- Ingestion API: http://localhost:8080/v1/events

### Option 2: Local Development (Without Docker)

**Prerequisites**: Go 1.21+, PHP 8.2+, Node.js 20+, Redis

#### Step 1: Start Redis
```bash
redis-server
```

#### Step 2: Start Go Ingestion Service
```bash
cd go-ingestion
go mod download
go run main.go
# Listening on :8080
```

#### Step 3: Start Go Consumer Service
```bash
cd go-consumer
go mod download
mkdir -p ../data
DB_PATH=../data/analytics.db go run main.go
# Consumer started, waiting for events...
```

#### Step 4: Start PHP Analytics Service
```bash
cd php-analytics
composer install
DB_PATH=../data/analytics.db php -S localhost:8000 -t public
# PHP development server started
```

#### Step 5: Start Frontend
```bash
cd frontend
npm install
VITE_API_URL=http://localhost:8000 npm run dev
# Vite dev server running on http://localhost:3000
```

## End-to-End Testing

### Test 1: Submit Events

```bash
# Test 1: Page View
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "page_view",
    "timestamp": "2025-10-02T10:00:00Z",
    "product_id": "prod-laptop-001"
  }'
# Expected: {"status":"accepted"}

# Test 2: More Page Views
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "page_view",
    "timestamp": "2025-10-02T10:01:00Z",
    "product_id": "prod-laptop-001"
  }'

curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "page_view",
    "timestamp": "2025-10-02T10:02:00Z",
    "product_id": "prod-phone-002"
  }'

# Test 3: Add to Cart
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "add_to_cart",
    "timestamp": "2025-10-02T10:05:00Z",
    "product_id": "prod-laptop-001"
  }'

# Test 4: Purchases
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "purchase",
    "timestamp": "2025-10-02T10:10:00Z",
    "product_id": "prod-laptop-001",
    "order_amount": 1299.99
  }'

curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "purchase",
    "timestamp": "2025-10-02T10:15:00Z",
    "product_id": "prod-phone-002",
    "order_amount": 899.99
  }'
```

### Test 2: Verify Consumer Processed Events

```bash
# Wait 5-10 seconds for consumer to batch and write events

# Check consumer logs
docker-compose logs consumer
# OR for local:
# Check terminal where consumer is running

# Expected: "Flushed X events to database"
```

### Test 3: Retrieve Analytics

```bash
curl http://localhost:8000/api/analytics | jq
```

**Expected Response:**
```json
{
  "total_page_views": 3,
  "total_add_to_carts": 1,
  "total_purchases": 2,
  "conversion_rate": 66.67,
  "average_purchase_value": 1099.99,
  "max_purchase_value": 1299.99,
  "min_purchase_value": 899.99,
  "top_product_id": "prod-laptop-001"
}
```

### Test 4: Verify Frontend Dashboard

1. Open http://localhost:3000 in browser
2. Should see dashboard with:
   - Total Page Views: 3
   - Total Add to Carts: 1
   - Total Purchases: 2
   - Conversion Rate: 66.67%
   - Average Order Value: $1099.99
   - Max Order Value: $1299.99
   - Min Order Value: $899.99
   - Top Product: prod-laptop-001

3. Dashboard auto-refreshes every 5 seconds

### Test 5: Health Checks

```bash
# Ingestion health
curl http://localhost:8080/v1/health
# Expected: {"status":"healthy"}

# Analytics health
curl http://localhost:8000/api/health
# Expected: {"status":"healthy"}
```

## Verification Checklist

- [ ] Redis running on port 6379
- [ ] Go Ingestion service responding on port 8080
- [ ] Go Consumer processing events (check logs)
- [ ] SQLite database created at `./data/analytics.db`
- [ ] PHP Analytics service responding on port 8000
- [ ] Frontend accessible on port 3000
- [ ] Events successfully ingested (202 Accepted)
- [ ] Events appear in database (check consumer logs)
- [ ] Analytics computed correctly
- [ ] Dashboard displays metrics
- [ ] Auto-refresh working

## Troubleshooting

### Port Already in Use
```bash
# Check what's using the port
lsof -i :6379  # Redis
lsof -i :8080  # Ingestion
lsof -i :8000  # Analytics
lsof -i :3000  # Frontend

# Kill process
kill -9 <PID>
```

### Docker Issues
```bash
# Clean everything
docker-compose down -v

# Rebuild from scratch
docker-compose build --no-cache

# Start fresh
docker-compose up -d
```

### Database Issues
```bash
# Remove database and restart
rm -f data/analytics.db
# Restart consumer service
```

### Go Module Issues
```bash
cd go-ingestion && go mod tidy
cd ../go-consumer && go mod tidy
```

### PHP Dependencies
```bash
cd php-analytics && composer install
```

### Frontend Issues
```bash
cd frontend && rm -rf node_modules && npm install
```

## Testing with api-examples.http

Use VS Code REST Client or IntelliJ HTTP Client:

1. Open `api-examples.http`
2. Click "Send Request" on any request
3. View response inline

## Performance Verification

```bash
# Submit 100 events
for i in {1..100}; do
  curl -s -X POST http://localhost:8080/v1/events \
    -H "Content-Type: application/json" \
    -d "{
      \"event_type\": \"page_view\",
      \"timestamp\": \"2025-10-02T10:00:00Z\",
      \"product_id\": \"prod-$i\"
    }" &
done
wait

# Check consumer batched them
docker-compose logs consumer | grep "Flushed"
# Should see: "Flushed 100 events to database"

# Verify analytics updated
curl http://localhost:8000/api/analytics | jq .total_page_views
# Should show increased count
```

## Database Inspection

```bash
# Install sqlite3 if needed
brew install sqlite3  # macOS
apt-get install sqlite3  # Ubuntu

# Connect to database
sqlite3 data/analytics.db

# Run queries
SELECT COUNT(*) FROM events;
SELECT event_type, COUNT(*) FROM events GROUP BY event_type;
SELECT * FROM events LIMIT 10;
.quit
```

## Production Readiness

This implementation is **demo-ready** but requires these changes for production:

1. **Replace SQLite with PostgreSQL**
2. **Replace Redis Pub/Sub with Kafka/Kinesis**
3. **Add authentication to all APIs**
4. **Implement rate limiting**
5. **Add monitoring (Prometheus/Grafana)**
6. **Deploy to AWS/GCP/Azure** (see docs/AWS-DEPLOYMENT.md)
7. **Set up CI/CD pipeline**
8. **Configure SSL/TLS**

## Support

If you encounter issues:

1. Check logs: `docker-compose logs <service>`
2. Verify all services are running: `docker-compose ps`
3. Review README.md for detailed documentation
4. Check docs/ARCHITECTURE.md for system design
5. Reach out in Slack: #gtk-case-study-oluwaseun-thani
