> Extracted from: admin-operations.md — Section: 1.11 Admin Plans

## 1.11 Admin Plans

- [ ] `GET /api/v1/admin/plans?limit=&offset=` — Paginated plan list with summaries (`AdminPlanHandlers.ListPlans` → `PlanRepo.GetAll`)
  - Paginates in-memory after fetch
  - Default `limit=20`, `offset=0`
- [ ] `GET /api/v1/admin/plans/{id}` — Plan detail with full limits and translations (`AdminPlanHandlers.GetPlan` → `PlanRepo.GetByID`)
- [ ] `PATCH /api/v1/admin/plans/{id}/limits` — Partial update of plan limits (`AdminPlanHandlers.UpdateLimits` → `PlanSvc.UpdatePlanLimits`)
  - JSON body with optional fields (pointer-based for partial updates)
  - Auto-invalidates cache after update
- [ ] `POST /api/v1/admin/plans/{id}/cache-invalidate` — Invalidate plan cache (`AdminPlanHandlers.InvalidateCache` → `PlanSvc.InvalidateCache`)

