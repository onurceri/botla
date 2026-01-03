# AGENTS.md - pkg/middleware

HTTP middleware stack for authentication, security, rate limiting, and request context propagation.

## OVERVIEW

Reusable HTTP middleware chain providing auth, CORS, rate limiting, security headers, request logging, and tenant context extraction.

## WHERE TO LOOK

- **Auth**: `auth.go` - JWT verification via header (`Authorization: Bearer`) or cookie (`botla_token`)
- **CORS**: `cors.go` - Deprecated wildcard version vs `CORSMiddlewareAllowOrigins` whitelist
- **Rate limiting**: `ratelimit.go` - Tiered approach: global (IP), plan-based (authenticated), endpoint overrides (auth endpoints)
- **Security headers**: `security.go` - CSP, HSTS, X-Frame-Options, nosniff
- **Request ID**: `request_id.go` - UUID generation + propagation via header/context
- **Request logging**: `requestlog.go` - Structured logging with status, bytes, duration, userID
- **Plan loader**: `plan_loader.go` - Loads user plan from DB into context (AFTER AuthMiddleware)
- **Organization**: `organization.go` - Tenant context via `X-Organization-ID` header + membership validation
- **Recovery**: `recovery.go` - Panic recovery with stack trace in dev, generic error in prod
- **Max bytes**: `maxbytes.go` - Request body size limit (DoS protection)

## CONVENTIONS

- **Middleware order** (defined in `cmd/server/main.go`):
  ```
  RequestID -> Security -> Recovery -> Logger -> MaxBytes -> PlanLoader -> RateLimit -> Router
  ```
- **Context pattern**: Helper functions like `UserIDFromContext(ctx)` for extraction
- **Rate limit headers**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After`
- **Context keys**: Private `contextKey` type in each file to avoid collisions
- **Plan loading**: Runs AFTER AuthMiddleware, BEFORE RateLimitMiddleware (needs plan for limits)

## ANTI-PATTERNS

- **Memory rate limiter**: `ratelimit.NewMemoryLimiter` is a fallback for dev/testing only. **NOT suitable for distributed deployments** - each instance has independent state, defeating rate limit consistency
- **PlanLoader placement**: Never place before AuthMiddleware - user ID required to load plan
- **Endpoint-specific rate limits**: Use only for sensitive endpoints (auth) - they bypass plan-based limits entirely
- **CORS wildcard**: `CORSMiddleware()` deprecated - production must use `CORSMiddlewareAllowOrigins` with explicit whitelist
