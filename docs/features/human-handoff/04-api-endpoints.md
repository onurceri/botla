> Extracted from: human-handoff.md — Section: 2. API Endpoints

# 2. API Endpoints

## 2.1 Handoff Endpoints

| Method | Path | Handler | Expected Status Codes |
|--------|------|---------|----------------------|
| POST | /api/v1/public/chatbots/{id}/handoff | HandoffHandler.Request | 201, 400, 404, 429 |
| POST | /api/v1/public/chatbots/{id}/handoff/{requestId}/contact | HandoffHandler.SubmitContact | 200, 400, 404, 409 |
| GET | /api/v1/chatbots/{id}/handoff-requests | HandoffHandler.List | 200, 401, 403, 404 |
| GET | /api/v1/chatbots/{id}/handoff-requests/{requestId} | HandoffHandler.Get | 200, 401, 403, 404 |
| PATCH | /api/v1/chatbots/{id}/handoff-requests/{requestId} | HandoffHandler.Update | 200, 400, 401, 403, 404, 422 |
