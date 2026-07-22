> Extracted from: admin-operations.md — Section: 1.2 Admin Users

## 1.2 Admin Users

- [ ] `GET /api/v1/admin/users?limit=&offset=&email=&is_platform_admin=&plan_id=` — Paginated user list with filters (`AdminHandlers.ListUsers` → `UserRepo.AdminListUsers`)
  - Default `limit=20`, `offset=0`
- [ ] `GET /api/v1/admin/users/{id}` — Get user detail (`AdminHandlers.GetUser` → `UserRepo.GetByID`)
- [ ] `PATCH /api/v1/admin/users/{id}` — Update user (JSON body, partial updates) (`AdminHandlers.UpdateUser` → `UserRepo.AdminUpdateUser`)
  - Logs action via `AdminService.LogAction`

