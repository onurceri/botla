> Extracted from: plans-entitlements.md — Section: 1.2 Plan Endpoints

## 1.2 Plan Endpoints

- [ ] `GET /api/v1/plans` — List all active plans (`PlansHandlers.GetAllPlans` → `PlanService.GetAllPlans`)
- [ ] `GET /api/v1/plans/{code}` — Get plan by code (`PlansHandlers.GetPlanByCode` → `PlanService.GetPlanByCode`)
- [ ] `GET /api/v1/me/plan` — Get current user's plan with features and available models (`PlanHandlers.GetPlan` → `PlanRepo.GetPlanWithLimits`)
- [ ] `GET /api/v1/me/usage` — Get current user's usage stats (`UsageHandlers.GetUsage`)

