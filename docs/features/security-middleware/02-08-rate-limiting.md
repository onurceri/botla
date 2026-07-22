> Extracted from: security-middleware.md — Section: 2.8 Rate Limiting (`ratelimit.go`)

## 2.8 Rate Limiting (`ratelimit.go`)

- [ ] Redis-backed sliding window implementation
- [ ] Per-plan rate limits are applied (different limits for free/pro/enterprise)
- [ ] Endpoint-specific overrides: chat, sources, auth login, auth register, auth refresh
- [ ] Global rate limits apply to all endpoints (safety net)
- [ ] Per-user rate limits apply to authenticated requests
- [ ] Per-IP rate limits apply to unauthenticated requests
- [ ] Rate limit headers are returned: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- [ ] `Retry-After` header is returned on rate limit exceeded (429)
- [ ] Rate limit state is stored in Redis (sliding window)
- [ ] Redis connection failure causes rate limiter to fail open (allow) or fail closed (deny) based on configuration
