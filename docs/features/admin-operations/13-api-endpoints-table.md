> Extracted from: admin-operations.md — Section: 2. API Endpoints Table

# 2. API Endpoints Table

| Method | Path | Handler | Auth | Success Codes | Error Codes |
|--------|------|---------|------|---------------|-------------|
| GET | `/api/v1/admin/stats/overview` | `ah.GetOverviewStats` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/users` | `ah.ListUsers` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/users/{id}` | `ah.GetUser` | JWT + Admin | 200 | 401, 403, 404, 500 |
| PATCH | `/api/v1/admin/users/{id}` | `ah.UpdateUser` | JWT + Admin | 200 | 401, 403, 400, 500 |
| GET | `/api/v1/admin/organizations` | `ah.ListOrganizations` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/organizations/{id}` | `ah.GetOrganization` | JWT + Admin | 200 | 401, 403, 404, 500 |
| GET | `/api/v1/admin/chatbots` | `ach.ListChatbots` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/chatbots/{id}` | `ach.GetChatbot` | JWT + Admin | 200 | 401, 403, 404, 500 |
| POST | `/api/v1/admin/chatbots/{id}/force-refresh` | `ach.ForceRefreshChatbot` | JWT + Admin | 200 | 401, 403, 404, 500 |
| GET | `/api/v1/admin/sources` | `ash.ListSources` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/sources/stats` | `ash.GetSourceStats` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/sources/types` | `ash.GetSourceTypes` | JWT + Admin | 200 | 401, 403 |
| GET | `/api/v1/admin/sources/{id}` | `ash.GetSource` | JWT + Admin | 200 | 401, 403, 404, 500 |
| POST | `/api/v1/admin/sources/{id}/reprocess` | `ash.ReprocessSource` | JWT + Admin | 200 | 401, 403, 404, 500 |
| GET | `/api/v1/admin/health/detailed` | `adhh.GetDetailedHealth` | JWT + Admin | 200, 503 | 401, 403 |
| GET | `/api/v1/admin/queues` | `aqh.GetQueues` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/queues/stuck` | `aqh.GetStuckJobs` | JWT + Admin | 200 | 401, 403, 500 |
| POST | `/api/v1/admin/queues/{id}/retry` | `aqh.RetryJob` | JWT + Admin | 200 | 401, 403, 400, 500 |
| DELETE | `/api/v1/admin/queues/{id}` | `aqh.DeleteJob` | JWT + Admin | 200 | 401, 403, 400, 500 |
| GET | `/api/v1/admin/errors` | `aeh.ListErrors` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/errors/stats` | `aeh.GetErrorStats` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/errors/{id}` | `aeh.GetError` | JWT + Admin | 200 | 401, 403, 404, 500 |
| GET | `/api/v1/admin/audit-logs` | `aah.ListAuditLogs` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/privacy/requests` | `aph.ListPrivacyRequests` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/privacy/requests/{id}` | `aph.GetPrivacyRequest` | JWT + Admin | 200 | 401, 403, 404, 500 |
| GET | `/api/v1/admin/privacy/requests/{id}/download` | `aph.DownloadPrivacyExport` | JWT + Admin | 200 | 401, 403, 400, 404, 410, 500, 503 |
| GET | `/api/v1/admin/privacy/requests/{id}/download-url` | `aph.GetDownloadURL` | JWT + Admin | 200 | 401, 403, 400, 404, 410, 500, 503 |
| PATCH | `/api/v1/admin/privacy/requests/{id}` | `aph.ProcessPrivacyRequest` | JWT + Admin | 200 | 401, 403, 400, 404, 500 |
| POST | `/api/v1/admin/privacy/export/{userId}` | `aph.GenerateUserExport` | JWT + Admin | 200 | 401, 403, 400, 500 |
| GET | `/api/v1/admin/privacy/exports/{id}/download` | `aph.DownloadDataExport` | JWT + Admin | 200 | 401, 403, 400, 404, 410, 500, 503 |
| GET | `/api/v1/admin/plans` | `apnh.ListPlans` | JWT + Admin | 200 | 401, 403, 500 |
| GET | `/api/v1/admin/plans/{id}` | `apnh.GetPlan` | JWT + Admin | 200 | 401, 403, 404, 500 |
| PATCH | `/api/v1/admin/plans/{id}/limits` | `apnh.UpdateLimits` | JWT + Admin | 200 | 401, 403, 400, 404, 500 |
| POST | `/api/v1/admin/plans/{id}/cache-invalidate` | `apnh.InvalidateCache` | JWT + Admin | 200 | 401, 403, 404, 500 |

---

