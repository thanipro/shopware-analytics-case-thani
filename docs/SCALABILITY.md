# Scalability

How I'd handle growth from demo to real traffic.

## Current Bottlenecks

Right now this handles maybe 5,000 events per second on a single Go instance. The main limits:

SQLite can't handle concurrent writes well - it might break around 1,000 writes/second. The in-memory queue works but if the container restarts, queued events are gone. And there's only one backend instance, so no horizontal scaling yet.

For analytics, PHP can handle maybe 100 concurrent requests before it starts slowing down. And there's no read scaling since everything hits the same database.

## First Step: 50K Events/Second

**Switch the database.** PostgreSQL instead of SQLite. It handles concurrent writes way better and you can add read replicas.

**Run multiple backend instances.** Put 5-10 containers behind a load balancer. Each has its own queue. ECS or Kubernetes can auto-scale based on CPU.

**Add Kafka.** Replace in-memory channels with a proper message queue. If a consumer crashes, Kafka keeps the events. You can replay them for debugging. Multiple consumers can process the queue in parallel.

**Read replicas for analytics.** Since the PHP service only reads data, point it at 2-3 PostgreSQL read replicas. This spreads the query load and keeps the primary database free for writes.

## Bigger Scale: 500K Events/Second

At this level, the architecture changes:

**Ingestion:** API Gateway + Lambda works well here because it auto-scales to thousands of instances. Or stick with Go but run 50+ containers.

**Queue:** Kafka  to handle the event stream. Keep events for 30 days so you can reprocess if needed.


**Caching:** Add Redis with a 10-second TTL for analytics results. This means the PHP API barely touches the database - most requests hit cache.

## Caching

Instead of calculating metrics on every request, cache them. Check Redis first, return cached data if it's there. Only recalculate every 10 seconds when cache expires.

This cuts database load by 90%+.

## Database Partitioning

For long-term storage, partition tables by month. Queries only scan relevant partitions. Archive old partitions to S3 for compliance but don't query them.


## Load Testing

Before going live with any changes, load test with tools like k6. Hit the backend with 10K events/second for a minute and watch:

- Does latency stay under 100ms?
- Is error rate still 0%?
- Is the queue filling up?
- Are database connections maxing out?

These tell you where the next bottleneck is.

## What Really Matters

Scaling comes down to:
1. Find the bottleneck (usually database or single instance)
2. Fix that one thing
3. Test again to find the next bottleneck
4. Repeat

Don't over-engineer for traffic you don't have yet. Scale when you need to, not before.
