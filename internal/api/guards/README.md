# API Guards

This package contains **API-specific authorization guards** that enforce access control rules for specific route groups.

## Contents

| File | Purpose |
|------|---------|
| `admin.go` | Validates platform admin access for `/admin/*` routes |

## When to Use This Package

Add guards here when:
- It's specific to the HTTP API layer
- It enforces authorization rules (not authentication)
- It applies to specific route groups, not globally

## When to Use `pkg/middleware/` Instead

Use `pkg/middleware/` for:
- Cross-cutting concerns (logging, CORS, rate limiting)
- Reusable middleware that could apply to any HTTP handler
- Authentication (JWT validation)
- Request enrichment (request ID, plan loading)
