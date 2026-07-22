> Extracted from: admin-operations.md — Section: 1.3 Admin Organizations

## 1.3 Admin Organizations

- [ ] `GET /api/v1/admin/organizations?limit=&offset=&name=&plan_id=` — Paginated org list with filters (`AdminHandlers.ListOrganizations` → `OrganizationRepo.AdminList`)
  - Default `limit=20`, `offset=0`
- [ ] `GET /api/v1/admin/organizations/{id}` — Get org detail (`AdminHandlers.GetOrganization` → `OrganizationRepo.GetByID`)

