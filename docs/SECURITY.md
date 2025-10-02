# Security & Privacy

What's missing and what I'd add for production.

## Current Gaps

Right now there's no authentication - anyone can submit events or read analytics. No rate limiting either, so it could get hammered. CORS allows all origins which is fine for a demo but not for production.

The database is shared between services, so technically the analytics API has write access even though it only needs read. And parameterized queries prevent SQL injection, but there's no validation beyond basic types.

## Authentication

For the backend, I'd use API keys. Each client gets a unique key to submit events. Store hashed versions in the database so even if it leaks, keys aren't exposed.

For the analytics API, JWT tokens or session-based auth makes sense. Only authenticated users should see the metrics.

## Rate Limiting

Without limits, someone could flood the backend with junk events or scrape the analytics endpoint.

For Go, a simple token bucket limiter works - allow X requests per second with some burst capacity. For multiple backend instances, use Redis to share the rate limit state.


## Input Validation

Already checking event types and required fields. I'd add:

- Reject really old timestamps (prevents log injection)
- Limit field lengths (product ID shouldn't be 10KB)

## CORS

Change from wildcard to a list of allowed origins. Only your actual frontend domains should be able to call the API.

## Data Privacy

Currently storing product IDs and amounts - no personal data. If user tracking gets added:

- Hash user IDs before storing them
- Don't log IP addresses
- Provide a deletion endpoint for GDPR compliance
- Set up data retention - delete events after 90 days or aggregate them and delete the raw data

## Database Security

The analytics service should use a read-only database user. The backend needs insert access but analytics only needs select.


## Security Headers

Add standard security headers to responses - prevents clickjacking, XSS, and other common attacks. Most frameworks make this easy to configure.

## Dependency Updates

Set up Dependabot or Renovate to auto-create PRs for dependency updates. Review and merge them regularly. Old dependencies are how most breaches happen.


## What Really Matters

Security basics:
1. Authenticate all requests
2. Limit request rates
3. Validate all input
4. Use HTTPS everywhere
5. Keep secrets out of code
6. Log security events
7. Update dependencies

Get these right and you're ahead of 90% of issues. Everything else is defense in depth.
