> Extracted from: actions-smart-tools.md — Section: 2. API Endpoints

# 2. API Endpoints

## 2.1 Action Endpoints

| Method | Path | Handler | Expected Status Codes |
|--------|------|---------|----------------------|
| GET | /api/v1/chatbots/{id}/actions | ActionHandler.List | 200, 401, 403, 404 |
| POST | /api/v1/chatbots/{id}/actions | ActionHandler.Create | 201, 400, 401, 403, 404, 422 |
| GET | /api/v1/chatbots/{id}/actions/{actionId} | ActionHandler.Get | 200, 401, 403, 404 |
| PUT | /api/v1/chatbots/{id}/actions/{actionId} | ActionHandler.Update | 200, 400, 401, 403, 404, 422 |
| DELETE | /api/v1/chatbots/{id}/actions/{actionId} | ActionHandler.Delete | 204, 401, 403, 404 |
| GET | /api/v1/chatbots/{id}/actions/logs | ActionHandler.Logs | 200, 401, 403, 404 |
