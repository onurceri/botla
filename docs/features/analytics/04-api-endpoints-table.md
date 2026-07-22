> Extracted from: analytics.md — Section: 2. API Endpoints Table

# 2. API Endpoints Table

| Method | Path | Handler | Auth | Success Codes | Error Codes |
|--------|------|---------|------|---------------|-------------|
| GET | `/api/v1/analytics` | `anh.GetAnalytics` | JWT + DeletedAccount + Tenant | 200 | 401, 403, 500 |
| GET | `/api/v1/chatbots/{id}/analytics/overview` | `anh.GetChatbotAnalyticsOverview` | JWT + Tenant | 200 | 401, 403, 404, 500 |
| GET | `/api/v1/chatbots/{id}/analytics/trends?days=N` | `anh.GetChatbotAnalyticsTrends` | JWT + Tenant | 200 | 401, 403, 404, 500 |
| GET | `/api/v1/chatbots/{id}/analytics/sources?days=N` | `anh.GetSourceUsage` | JWT + Tenant | 200 | 401, 403, 404, 500 |

---

