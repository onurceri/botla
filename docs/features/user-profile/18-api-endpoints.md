> Extracted from: user-profile.md — Section 1.3

# 1.3 API Endpoints Table

| Method | Path | Handler | Expected Status Codes |
|--------|------|---------|----------------------|
| `GET` | `/api/v1/me` | `mh.Me` | 200, 401 |
| `GET` | `/api/v1/me/plan` | `plh.GetPlan` | 200, 401 |
| `GET` | `/api/v1/me/usage` | `uh.GetUsage` | 200, 401 |
| `GET` | `/api/v1/me/onboarding` | `oh.Get` | 200, 401 |
| `PUT` | `/api/v1/me/onboarding` | `oh.Update` | 200, 400, 401 |
| `POST` | `/api/v1/me/onboarding/skip` | `oh.Skip` | 200, 401 |
| `POST` | `/api/v1/me/onboarding/complete` | `oh.Complete` | 200, 401 |
| `GET` | `/api/v1/me/privacy/consents` | `uph.List` | 200, 401 |
| `PATCH` | `/api/v1/me/privacy/consents` | `uph.Update` | 200, 400, 401 |
| `POST` | `/api/v1/me/privacy/export` | `uph.RequestExport` | 201, 400, 401, 409 |
| `POST` | `/api/v1/me/privacy/correction` | `uph.RequestCorrection` | 201, 400, 401 |
| `POST` | `/api/v1/me/privacy/delete` | `uph.RequestDeletion` | 201, 400, 401, 409 |
| `GET` | `/api/v1/me/privacy/requests` | `uph.ListRequests` | 200, 401 |
| `GET` | `/api/v1/me/privacy/requests/{id}` | `uph.GetRequest` | 200, 401, 404 |
| `DELETE` | `/api/v1/me/privacy/requests/{id}` | `uph.DeleteRequest` | 200, 401, 404, 409 |
| `GET` | `/api/v1/me/privacy/requests/{id}/download` | `uph.DownloadExport` | 200, 401, 404 |
| `GET` | `/api/v1/me/privacy/exports/{id}/download` | `uph.DownloadExport` | 200, 401, 404 |
