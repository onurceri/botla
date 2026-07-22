> Extracted from: admin-operations.md — Section: 1.9 Admin Audit

## 1.9 Admin Audit

- [ ] `GET /api/v1/admin/audit-logs?limit=&offset=&admin_user_id=&action=&target_type=&start_date=&end_date=` — Paginated audit logs (`AdminAuditHandlers.ListAuditLogs` → `AdminRepo.ListAuditLogs`)
  - Default `limit=50`, `offset=0`
  - Filters: `admin_user_id`, `action`, `target_type`, `start_date` (RFC3339), `end_date` (RFC3339)
  - Response includes `data`, `total`, `page`, `per_page`

