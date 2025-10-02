# Monitoring & Alerting

What I'd watch and how I'd know when things break.

## What to Monitor

For the backend, I care about request rate, how full the event queue is getting, and whether batch writes to the database are failing or taking too long.

For the analytics API, it's mainly request latency and error rates. If queries start taking multiple seconds, something's wrong.

Database metrics matter too - connection count, query performance, and disk space. SQLite won't tell you much, but with PostgreSQL you get detailed stats.

Infrastructure basics: CPU, memory, disk I/O. If CPU stays above 70-80% consistently, time to scale.

## How I'd Set It Up

### Prometheus + Grafana

This is the standard choice for metrics. Backend exposes a `/metrics` endpoint, Prometheus scrapes it every 15 seconds, Grafana shows dashboards.

I'd track things like events received, queue depth, and batch flush duration. Set up alerts for when queue depth hits 80% or error rate goes above 5%.

Grafana dashboards make it easy to spot patterns - you can see if traffic is spiking, if the queue is building up, or if database writes are slowing down.

### CloudWatch (if on AWS)

Honestly, if you're already on AWS, just use CloudWatch. It's built-in and works fine.

Set up log groups for each service, create alarms for high error rates or resource usage, and send notifications to SNS. SNS can email you or hit a Slack webhook.

Less setup than Prometheus, and it's one less thing to maintain.

## Alerts

I'd keep alerts simple and actionable:

- Queue nearly full → might need to scale up or investigate slow database
- Error rate above 5% for more than 2 minutes → something's broken
- Database writes taking >2 seconds → check database load or lock contention
- CPU above 80% for 5+ minutes → time to add capacity

Send critical alerts to Slack or PagerDuty. Send warnings to email. Too many alerts and people start ignoring them.

## When Things Break

First place I'd check: logs. Go service logs show batch write failures. PHP logs show query errors or exceptions.

Next: metrics. Is the queue full? Are requests timing out? Is the database responding?

Then: database itself. Check active connections, look for slow queries, make sure there's disk space.

Usually it's one of three things: traffic spike you didn't expect, database getting hammered, or a bug in the new code.

## Health Checks

Both services need a `/health` endpoint. Keep it simple - ping the database, return 200 if it responds, 500 if not.

The load balancer uses this to know which containers are healthy. If a container fails health checks, it gets replaced automatically.

## Logging

Structured logs make debugging way easier. Instead of random strings, log JSON with fields like level, message, error, and relevant IDs.

For AWS, ship logs to CloudWatch. For on-prem or if you want more control, ELK stack works great. For low traffic, even just grepping local log files is fine.

## What Actually Matters

Good monitoring comes down to:
1. Know when error rates spike
2. Know when things are slow
3. Know when you're running out of resources
4. Be able to check logs quickly

Everything else is nice to have but not essential.
