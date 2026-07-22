> Extracted from: plans-entitlements.md — Section: 2. API Endpoints Table

# 2. API Endpoints Table

| Method | Path | Handler | Auth | Success Codes | Error Codes |
|--------|------|---------|------|---------------|-------------|
| GET | `/api/v1/plans` | `plansH.GetAllPlans` | Public | 200 | 500 |
| GET | `/api/v1/plans/{code}` | `plansH.GetPlanByCode` | Public | 200 | 400, 404, 500 |
| GET | `/api/v1/me/plan` | `plh.GetPlan` | JWT + DeletedAccount | 200 | 401, 404, 500 |
| GET | `/api/v1/me/usage` | `uh.GetUsage` | JWT + DeletedAccount + Tenant | 200 | 401, 500 |

---

