> Extracted from: admin-operations.md — Section: 1.10 Admin Privacy (KVKK)

## 1.10 Admin Privacy (KVKK)

- [ ] `GET /api/v1/admin/privacy/requests?status=&page=&limit=` — List privacy requests (`PrivacyHandlers.ListPrivacyRequests` → `PrivacyRepo.ListPrivacyRequests`)
  - Default `page=1`, `limit=20` (max 100)
  - Status filter: `pending`, `processing`, `completed`, `denied`
- [ ] `GET /api/v1/admin/privacy/requests/{id}` — Get request detail (`PrivacyHandlers.GetPrivacyRequest` → `PrivacyRepo.GetPrivacyRequest`)
- [ ] `GET /api/v1/admin/privacy/requests/{id}/download` — Download export file (`PrivacyHandlers.DownloadPrivacyExport` → `Storage.DownloadFile`)
  - Validates request is `export` type and `completed` status
  - Checks export URL expiry (returns 410 Gone if expired)
- [ ] `GET /api/v1/admin/privacy/requests/{id}/download-url` — Get signed download URL (`PrivacyHandlers.GetDownloadURL` → `Storage.GetSignedURL`)
  - Returns signed URL with 1-hour expiry
- [ ] `PATCH /api/v1/admin/privacy/requests/{id}` — Process request (`PrivacyHandlers.ProcessPrivacyRequest`)
  - Body: `{"action": "approve"|"deny", "denial_reason": "..."}`
  - Approve: handles `deletion` → `PrivacyService.ProcessDeletion`, `export` → `PrivacyService.ProcessExportRequest`, others → mark completed
  - Deny: marks as `denied` with optional reason
  - Logs action via `AdminService.LogAction`
- [ ] `POST /api/v1/admin/privacy/export/{userId}` — Generate user data export (`PrivacyHandlers.GenerateUserExport` → `PrivacyService.ExportUserData`)
  - Logs action via `AdminService.LogAction`
- [ ] `GET /api/v1/admin/privacy/exports/{id}/download` — Download data export (`PrivacyHandlers.DownloadDataExport` → `Storage.DownloadFile`)
  - Validates export is `completed` status and not expired

