# Testing Strategy

How I'd make sure this works reliably.

## What's Tested Now

The Go backend has unit tests for the HTTP handlers - checking that valid events get accepted, invalid ones get rejected, and the queue-full scenario is handled.

PHP has unit tests for the analytics service using mocked repositories. Tests the calculation logic like conversion rates.

Frontend is just manually tested for now. Nothing automated.

Missing: integration tests where services actually talk to each other, load tests to see where it breaks, and any database migration tests.

## Test Pyramid

Most tests should be unit tests - fast, focused, test one thing. Some integration tests to make sure services work together. A few end-to-end tests for critical user flows.

The ratio I'd aim for: 80% unit, 15% integration, 5% end-to-end.

## Integration Tests

These would test the actual flow: submit an event via the API, wait for batch processing, check it's in the database. Then query analytics and verify the metrics match.

For Go, spin up a real database in memory, start the consumer, send events, wait 6 seconds for the batch to flush, then query to verify.

For PHP, same idea - use an in-memory SQLite database, insert test data, run the analytics service, check the results match what you expect.

## End-to-End Tests

Would use something like Playwright or Cypress to test the full user journey:
1. Submit events via the API
2. Wait for processing
3. Open the frontend dashboard
4. Verify the metrics show up correctly

This catches integration issues and frontend bugs you'd miss with unit tests.

## Load Testing

Before deploying any scaling changes, I'd load test with k6 or vegeta. Throw 10K events per second at the backend for a minute and watch what breaks.

Key metrics:
- Does latency stay reasonable (under 100ms)?
- Any errors?
- Is the queue filling up?
- Are database connections maxing out?

This tells you where the bottleneck actually is, not where you think it is.

## Manual Testing Checklist

Before any deployment:
- Submit each event type via curl
- Verify analytics API returns correct data
- Open dashboard and check the display
- Test error cases (invalid event type, malformed JSON)
- Submit 1000 events and verify batch processing
- Restart the backend and make sure it recovers

Takes 10 minutes and catches most regressions.

## Test Coverage

I'd aim for >80% coverage on backend code and >90% on the analytics service (since it's mostly business logic). Frontend >70% once we add tests there.

But coverage percentage doesn't mean much - what matters is testing the important paths and edge cases.

## What I'd Add

- Frontend component tests with Vitest
- Contract tests to ensure API matches what frontend expects
- Chaos testing - randomly kill services and see if the system recovers
- Database migration rollback tests

For now though, solid unit tests and a few integration tests are enough.

## Running Tests in CI

Every push should run all tests. If any fail, block the merge. Simple rule that prevents most bugs from reaching production.

Check coverage too - if it drops below the threshold, fail the build. Keeps people from skipping tests.
