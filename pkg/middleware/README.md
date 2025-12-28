# Infrastructure Middleware

This package contains **reusable, cross-cutting middleware** that applies globally or to broad categories of routes.

## Contents

| File | Purpose |
|------|---------|
| `auth.go` | JWT authentication |
| `cors.go` | CORS handling |
| `ratelimit.go` | Rate limiting |
| `security.go` | Security headers (CSP, HSTS, etc.) |
| `request_id.go` | Request ID propagation |
| `recovery.go` | Panic recovery |
| `requestlog.go` | Request logging |
| `plan_loader.go` | Load user plan into context |
| `plan_context.go` | Plan context helpers |
| `maxbytes.go` | Request size limits |
| `organization.go` | Tenant context extraction |

## When to Use This Package

Add middleware here when:
- It's a cross-cutting concern (applies broadly)
- It's reusable across different applications
- It handles authentication (not authorization)
- It enriches requests with context

## Middleware Chain Order

The recommended order is defined in `cmd/server/main.go`:

```
RequestID -> Security -> Recovery -> Logger -> MaxBytes -> PlanLoader -> RateLimit -> Router
```

## See Also

For API-specific authorization guards, see `internal/api/guards/`.
