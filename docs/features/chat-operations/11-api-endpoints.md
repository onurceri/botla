> Extracted from: chat-operations.md — Section: 2. API Endpoints

# 2. API Endpoints

## 2.1 Chat Endpoints

| Method | Path | Handler | Expected Status Codes |
|--------|------|---------|----------------------|
| POST | /api/v1/chatbots/{id}/chat | ChatHandler.AuthenticatedChat | 200, 400, 401, 403, 404, 422, 429 |
| POST | /api/v1/public/chatbots/{id}/chat | ChatHandler.PublicChat | 200, 400, 404, 422, 429 |
| GET | /api/v1/public/chatbots/{id}/config | ChatHandler.PublicConfig | 200, 404 |
| POST | /api/v1/public/chatbots/{id}/feedback | ChatHandler.Feedback | 200, 400, 404 |
