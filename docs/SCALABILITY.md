# Scalability Considerations

## Current Architecture Limitations

### Bottlenecks
1. **SQLite**: Limited to ~1,000 concurrent writes/second
2. **Single consumer**: One instance processing all events
3. **No caching**: Every analytics request hits database
4. **Redis Pub/Sub**: Ephemeral (no persistence if consumer is down)
5. **Synchronous aggregations**: Computed on every request

### Current Capacity
- **Ingestion**: ~10,000 events/second (single instance)
- **Consumer**: ~1,000 events/second (SQLite limitation)
- **Analytics API**: ~100 requests/second (no caching)

## Scaling to Millions of Events/Day

### Scenario: 10 Million Events/Day

**Breakdown**:
- 10,000,000 events / 86,400 seconds = **~116 events/second average**
- Peak hours (assume 10x): **~1,160 events/second**

**Current system**: ✅ Can handle with minor optimizations

### Scenario: 100 Million Events/Day

**Breakdown**:
- 100,000,000 events / 86,400 seconds = **~1,157 events/second average**
- Peak hours (assume 10x): **~11,570 events/second**

**Current system**: ❌ Needs major architectural changes

---

## Scaling Strategy

### Phase 1: Optimize Current Architecture (100M events/day)

#### 1.1 Migrate to PostgreSQL
```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: analytics
      POSTGRES_USER: analytics
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

**Benefits**:
- Concurrent writes: 10,000+ events/second
- Connection pooling
- Replication support

#### 1.2 Add Read Replicas
```
┌──────────┐
│ Primary  │ ← Consumer writes here
│   DB     │
└────┬─────┘
     │ Replicate
     ▼
┌──────────┐
│ Replica  │ ← Analytics reads here
│   DB     │
└──────────┘
```

**Benefits**:
- Isolate read/write workloads
- Scale reads independently
- Reduce primary DB load

#### 1.3 Add Caching Layer
```go
// Analytics service with Redis cache
func (s *AnalyticsService) GetAnalytics() map[string]interface{} {
    cacheKey := "analytics:current"

    // Try cache first
    if cached := s.redis.Get(cacheKey); cached != nil {
        return cached
    }

    // Compute from database
    analytics := s.computeAnalytics()

    // Cache for 5 seconds
    s.redis.SetEx(cacheKey, analytics, 5*time.Second)

    return analytics
}
```

**Benefits**:
- Reduce database queries by 95%+
- Sub-millisecond response times
- Lower database CPU usage

**Result**: ✅ Handles 100M events/day

---

### Phase 2: Horizontal Scaling (1B events/day)

#### 2.1 Replace Redis Pub/Sub with Kafka

```yaml
services:
  kafka:
    image: confluentinc/cp-kafka:latest
    environment:
      KAFKA_NUM_PARTITIONS: 10
      KAFKA_REPLICATION_FACTOR: 3
```

**Why Kafka?**
- Persistent storage (events not lost)
- Partitioning for parallel processing
- Message replay capability
- Handles millions of messages/second

**Architecture**:
```
Ingestion → Kafka (10 partitions) → 10 Consumer instances
```

#### 2.2 Scale Ingestion Horizontally
```
       ┌──────────────┐
       │Load Balancer │
       └──────┬───────┘
              │
    ┌─────────┼─────────┐
    ▼         ▼         ▼
┌────────┐┌────────┐┌────────┐
│Ingest 1││Ingest 2││Ingest N│
└────────┘└────────┘└────────┘
```

**Kubernetes Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ingestion
spec:
  replicas: 5
  template:
    spec:
      containers:
      - name: ingestion
        image: analytics/ingestion:latest
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
```

**Auto-scaling**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ingestion-hpa
spec:
  scaleTargetRef:
    kind: Deployment
    name: ingestion
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

#### 2.3 Scale Consumer with Consumer Groups
```go
// Consumer group configuration
config := kafka.ConsumerConfig{
    GroupID: "analytics-consumers",
    Topics:  []string{"events"},
}

// Each consumer processes different partitions
consumer := kafka.NewConsumer(config)
```

**Architecture**:
```
Kafka Partition 0 → Consumer 1
Kafka Partition 1 → Consumer 2
Kafka Partition 2 → Consumer 3
...
Kafka Partition 9 → Consumer 10
```

**Benefits**:
- Parallel processing
- Automatic partition rebalancing
- Fault tolerance

**Result**: ✅ Handles 1B events/day

---

### Phase 3: Time-Series Optimization (10B+ events/day)

#### 3.1 Migrate to ClickHouse

**Why ClickHouse?**
- Columnar storage (analytics-optimized)
- Compression: 10-100x smaller storage
- Aggregations: 100-1000x faster
- Handles billions of rows easily

**Schema**:
```sql
CREATE TABLE events (
    event_type String,
    timestamp DateTime,
    product_id String,
    order_amount Nullable(Float64),
    date Date DEFAULT toDate(timestamp)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (event_type, timestamp, product_id);
```

**Benefits**:
- Query 1B rows in milliseconds
- Automatic data partitioning
- Built-in materialized views

#### 3.2 Pre-compute Aggregations

**Materialized View**:
```sql
CREATE MATERIALIZED VIEW analytics_hourly
ENGINE = SummingMergeTree()
ORDER BY (hour, event_type)
AS SELECT
    toStartOfHour(timestamp) as hour,
    event_type,
    count() as count,
    avg(order_amount) as avg_amount,
    max(order_amount) as max_amount,
    min(order_amount) as min_amount
FROM events
GROUP BY hour, event_type;
```

**Benefits**:
- Real-time aggregations
- No full table scans
- Constant query time regardless of data volume

#### 3.3 Data Lifecycle Management

```sql
-- Partition retention (keep 90 days)
ALTER TABLE events DROP PARTITION '202501';

-- Move old data to cold storage (S3)
ALTER TABLE events MOVE PARTITION '202503' TO VOLUME 's3_cold';
```

**Storage Tiers**:
- **Hot** (SSD): Last 7 days - 10ms queries
- **Warm** (HDD): 8-30 days - 100ms queries
- **Cold** (S3): 31-90 days - 1s queries
- **Archive** (S3 Glacier): 90+ days - minutes

**Result**: ✅ Handles 10B+ events/day

---

## Infrastructure Scaling (AWS Example)

### Small Scale (100M events/day)
```
- Ingestion: 2x t3.medium (ECS Fargate)
- Kafka: 3x t3.large (MSK cluster)
- Consumer: 3x t3.medium (ECS Fargate)
- Database: db.r5.xlarge (RDS PostgreSQL)
- Cache: cache.r5.large (ElastiCache Redis)
- Analytics: 2x t3.small (ECS Fargate)

Estimated cost: ~$800/month
```

### Medium Scale (1B events/day)
```
- Ingestion: 5-20x t3.medium (auto-scaling)
- Kafka: 6x r5.xlarge (MSK)
- Consumer: 10x t3.large (auto-scaling)
- Database: db.r5.4xlarge + 2 read replicas
- Cache: cache.r5.xlarge (cluster mode)
- Analytics: 5x t3.medium (auto-scaling)

Estimated cost: ~$5,000/month
```

### Large Scale (10B events/day)
```
- Ingestion: 20-100x c5.2xlarge (auto-scaling)
- Kafka: 12x r5.2xlarge (MSK)
- Consumer: 50x c5.xlarge (auto-scaling)
- ClickHouse: 6x r5.8xlarge (self-managed cluster)
- Cache: cache.r6g.2xlarge (cluster mode)
- Analytics: 10x c5.large (auto-scaling)

Estimated cost: ~$30,000/month
```

---

## Query Optimization

### Current (PostgreSQL)
```sql
-- Slow for billions of rows
SELECT event_type, COUNT(*) as count
FROM events
GROUP BY event_type;
```

### Optimized (Pre-aggregated)
```sql
-- Real-time table updated by triggers/streams
CREATE TABLE event_counts (
    event_type VARCHAR(50),
    count BIGINT,
    last_updated TIMESTAMP,
    PRIMARY KEY (event_type)
);

-- Increment on insert
INSERT INTO event_counts (event_type, count, last_updated)
VALUES ('page_view', 1, NOW())
ON CONFLICT (event_type)
DO UPDATE SET
    count = event_counts.count + 1,
    last_updated = NOW();
```

**Query time**: O(n) → O(1)

---

## Stream Processing Architecture

### Lambda Architecture

```
┌─────────────────┐
│  Speed Layer    │  Real-time aggregations (Redis)
│  (Kafka Streams)│  Last 5 minutes
└────────┬────────┘
         │
         ├──────────────┐
         │              │
┌────────▼────────┐ ┌──▼─────────────┐
│  Batch Layer    │ │  Serving Layer │
│  (ClickHouse)   │ │  (API)         │
│  Historical     │ │  Merges both   │
└─────────────────┘ └────────────────┘
```

**Speed Layer** (Kafka Streams):
```java
stream.groupBy(event -> event.type)
      .windowedBy(TimeWindows.of(Duration.ofMinutes(5)))
      .count()
      .toStream()
      .to("real-time-counts");
```

**Benefits**:
- Real-time metrics (5-second latency)
- Historical accuracy
- Fault tolerance

---

## Cost Optimization

### 1. Tiered Storage
- Hot data (7 days): SSD
- Warm data (30 days): HDD
- Cold data (90 days): S3
- Archive (1 year): S3 Glacier

**Savings**: 70-90% storage costs

### 2. Reserved Instances
- 1-year commitment: 30% discount
- 3-year commitment: 50% discount

**Savings**: $2,000/month → $1,000/month

### 3. Spot Instances (for consumers)
- Use for stateless consumers
- 70-90% cheaper than on-demand

**Savings**: $500/month → $100/month

### 4. Data Compression
- ClickHouse: 10-100x compression
- Parquet (S3): 5-10x compression

**Savings**: 90% storage costs

---

## Performance Targets

| Metric | Current | Phase 1 | Phase 2 | Phase 3 |
|--------|---------|---------|---------|---------|
| Ingestion throughput | 1K/s | 10K/s | 100K/s | 1M/s |
| Query latency (p95) | 100ms | 50ms | 10ms | 5ms |
| Data retention | 30 days | 90 days | 1 year | Unlimited |
| Analytics lag | 5s | 5s | 1s | 100ms |
| Cost per 1M events | $10 | $5 | $2 | $0.50 |

---

## Monitoring & Alerting

### Key Metrics

**Ingestion**:
- Events per second
- Error rate
- p95/p99 latency

**Queue**:
- Message lag
- Partition distribution
- Consumer group lag

**Database**:
- Query latency
- Connection pool usage
- Replication lag

**Alerts**:
```yaml
- name: IngestionHighErrorRate
  condition: error_rate > 1%
  action: PagerDuty

- name: ConsumerLagHigh
  condition: lag > 100000 messages
  action: Auto-scale consumers

- name: DatabaseHighCPU
  condition: cpu > 80%
  action: Scale read replicas
```

---

## Summary

| Scale | Events/Day | Architecture | Key Changes |
|-------|-----------|--------------|-------------|
| **Current** | 1M | SQLite + Redis Pub/Sub | None |
| **Phase 1** | 100M | PostgreSQL + Redis | Caching, read replicas |
| **Phase 2** | 1B | Kafka + Horizontal scaling | Multiple consumers, partitioning |
| **Phase 3** | 10B+ | ClickHouse + Stream processing | Columnar storage, materialized views |

**Key Takeaway**: Start simple, scale incrementally based on actual traffic patterns.
