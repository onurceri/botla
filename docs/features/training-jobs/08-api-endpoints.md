> Extracted from: training-jobs.md — Section: 3. API Endpoints Table

# 3. API Endpoints Table

| Method | Path | Handler | Success Codes | Error Codes |
|--------|------|---------|---------------|-------------|
| GET | `/api/v1/sources/{id}/job` | `GetJobStatus` | 200 | 401, 404, 500 |
| POST | `/api/v1/sources/{id}/job/retry` | `RetryJob` | 202 | 400, 401, 404, 500 |
| GET | `/api/v1/chatbots/{id}/pending-urls` | `ListPendingURLs` | 200 | 401, 404, 500 |
| POST | `/api/v1/chatbots/{id}/pending-urls/approve` | `ApprovePendingURLs` | 200 | 400, 401, 404, 500 |
| POST | `/api/v1/chatbots/{id}/pending-urls/reject` | `RejectPendingURLs` | 200 | 400, 401, 404, 500 |
| DELETE | `/api/v1/chatbots/{id}/pending-urls/clear` | `ClearPendingURLs` | 200 | 401, 404, 500 |
