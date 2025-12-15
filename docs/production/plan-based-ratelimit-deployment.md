# Plan-Based Rate Limiting - Production Deployment Guide

## Overview

The plan-based rate limiting feature has been implemented to provide different rate limits based on user subscription plans (Free, Pro, Ultra). This guide explains how to deploy the database migration to production.

## Database Migration

### Migration File

- **File**: `db/migrations/000033_add_rate_limits_to_plans.up.sql`
- **Purpose**: Adds `rate_limits` configuration to existing Free and Pro plans, and creates the Ultra plan

### Rate Limits by Plan

| Plan  | Global Limit (req/min) | Chat Endpoint (req/min) | Sources Endpoint (req/min) |
|-------|------------------------|-------------------------|----------------------------|
| Free  | 100                    | 30                      | 10                         |
| Pro   | 500                    | 100                     | 30                         |
| Ultra | 2000                   | 500                     | 100                        |

## Running Migration on Render.com

Since Render.com free tier doesn't allow SSH access, you'll need to run the migration using the `migrate` CLI tool locally, connecting to the production database.

### Prerequisites

1. Install the `migrate` CLI tool:
   ```bash
   # macOS
   brew install golang-migrate
   
   # Or download from: https://github.com/golang-migrate/migrate/releases
   ```

2. Get your production database connection string from Render.com:
   - Go to your Render dashboard
   - Navigate to your PostgreSQL database
   - Copy the "External Database URL"

### Running the Migration

```bash
# Set the production database URL
export DATABASE_URL="postgresql://user:password@host:port/dbname?sslmode=require"

# Run the migration
migrate -path=db/migrations -database=$DATABASE_URL up
```

### Verification

After running the migration, verify that the rate limits were added:

```bash
# Connect to the production database
psql $DATABASE_URL

# Check the rate_limits configuration
SELECT code, config->'rate_limits' FROM plans WHERE code IN ('free', 'pro', 'ultra');
```

You should see output similar to:
```
 code  |                                                    ?column?                                                    
-------+---------------------------------------------------------------------------------------------------------------
 free  | {"endpoints": {"chat": {"window_seconds": 60, "requests_per_minute": 30}, ...}, "window_seconds": 60, ...}
 pro   | {"endpoints": {"chat": {"window_seconds": 60, "requests_per_minute": 100}, ...}, "window_seconds": 60, ...}
 ultra | {"endpoints": {"chat": {"window_seconds": 60, "requests_per_minute": 500}, ...}, "window_seconds": 60, ...}
```

### Rollback (if needed)

If you need to rollback the migration:

```bash
migrate -path=db/migrations -database=$DATABASE_URL down 1
```

## How It Works

### Architecture

1. **PlanLoaderMiddleware**: Loads the user's plan from the database on each authenticated request and stores it in the request context
2. **RateLimitMiddleware**: Uses the plan from context to create/retrieve a plan-specific rate limiter
3. **Dynamic Limiter Creation**: Rate limiters are created on-demand per plan code (not per user) to minimize memory usage

### Middleware Chain

```
Request → Auth → PlanLoader → RateLimit → Handler
```

### Key Features

- **Backward Compatible**: Falls back to global IP-based limits for unauthenticated users
- **Dynamic**: No code deployment needed to adjust limits (update database only)
- **Efficient**: 
  - Plan loaded once per request and cached in context
  - Limiters shared across all users of the same plan
  - Single database query via JOIN to fetch user + plan
- **Resilient**: Continues to work even if Redis is unavailable (falls back to in-memory)

### Performance Impact

- **Plan Loading**: ~1-2ms per request (single database JOIN query)
- **Memory**: One limiter instance per plan code (3 instances for Free/Pro/Ultra)
- **Redis Keys**: One key per active user (format: `ratelimit:user:{userID}`)

## Testing Locally

To test the plan-based rate limiting locally:

1. Apply the migration:
   ```bash
   make migrate-up
   ```

2. Start the server:
   ```bash
   make be-run
   ```

3. Create test users on different plans and verify rate limits work correctly.

## Monitoring

After deployment, monitor:

1. **Rate limit hit rates per plan**: Check X-RateLimit-* headers in responses
2. **Redis memory usage**: More limiters = more keys
3. **Plan lookup performance**: Check logs for `failed_to_load_plan` warnings
4. **429 responses**: Track which users/plans are hitting limits

## Future Enhancements

Potential improvements in the future:

1. **Endpoint-specific limits**: Use the `endpoints` config in plan to apply different limits per endpoint
2. **Admin UI**: Allow adjusting limits without database access
3. **Per-user overrides**: Special limits for VIP customers
4. **Soft limits**: Warning before hard cutoff
5. **Analytics dashboard**: Visualize usage patterns per plan

## Troubleshooting

### Users hitting limits unexpectedly

1. Check the user's plan: `SELECT plan_id FROM users WHERE id = 'user-id';`
2. Check plan configuration: `SELECT config->'rate_limits' FROM plans WHERE id = 'plan-id';`
3. Check Redis keys: `redis-cli KEYS "ratelimit:user:*"`

### Migration fails

1. Check database connection: `psql $DATABASE_URL -c "SELECT 1"`
2. Check migration status: `migrate -path=db/migrations -database=$DATABASE_URL version`
3. Review migration file for syntax errors

### Rate limits not applying

1. Check logs for `failed_to_load_plan` errors
2. Verify PlanLoaderMiddleware is in the middleware chain
3. Ensure user has a valid `plan_id` in the database

## Support

For issues or questions, refer to:
- Source code: `pkg/middleware/ratelimit.go`, `pkg/middleware/plan_loader.go`
- Migration: `db/migrations/000033_add_rate_limits_to_plans.up.sql`
- Plan model: `internal/models/plan.go`
