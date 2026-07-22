> Extracted from: security-middleware.md — Section: 2.7 Plan Middleware (`plan_context.go`, `plan_loader.go`)

## 2.7 Plan Middleware (`plan_context.go`, `plan_loader.go`)

- [ ] `PlanLoader` fetches the user's plan and injects it into context
- [ ] Plan context includes: plan ID, feature flags, rate limits, quota limits
- [ ] Caches plan data to avoid repeated DB queries
- [ ] Falls back to default (free) plan if plan is not found
- [ ] Invalid plan configuration in DB causes startup failure
- [ ] Plan data is available to downstream handlers and middlewares
