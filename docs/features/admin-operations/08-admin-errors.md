> Extracted from: admin-operations.md — Section: 1.8 Admin Errors

## 1.8 Admin Errors

- [ ] `GET /api/v1/admin/errors?limit=&offset=&severity=` — Paginated error list (`AdminErrorHandlers.ListErrors` → `AdminRepo.ListErrorLogs`)
  - Default `limit=20`, `offset=0`
  - Optional `severity` filter
- [ ] `GET /api/v1/admin/errors/stats` — Error summary stats (`AdminErrorHandlers.GetErrorStats` → `AdminRepo.GetErrorStats`)
- [ ] `GET /api/v1/admin/errors/{id}` — Error detail (`AdminErrorHandlers.GetError` → `AdminRepo.GetErrorLogByID`)

