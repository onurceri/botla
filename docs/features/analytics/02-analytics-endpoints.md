> Extracted from: analytics.md — Section: 1.2 Analytics Endpoints

## 1.2 Analytics Endpoints

- [ ] `GET /api/v1/analytics` — Global analytics across user's scope (`AnalyticsHandlers.GetAnalytics` → `AnalyticsRepo.GetGlobalAnalytics`)
  - Supports workspace-scoped and org-scoped context via `ExtractTenantContext`
  - Validates workspace membership via `WorkspaceService.GetWorkspace` + `OrgService.CheckMembership`
  - Validates org membership for global org context
- [ ] `GET /api/v1/chatbots/{id}/analytics/overview` — Chatbot analytics overview (`GetChatbotAnalyticsOverview` → `AnalyticsService.GetChatbotOverview`)
- [ ] `GET /api/v1/chatbots/{id}/analytics/trends?days=N` — Daily trend data (`GetChatbotAnalyticsTrends` → `AnalyticsService.GetChatbotTrends`)
  - Default `days=30`, max 365
- [ ] `GET /api/v1/chatbots/{id}/analytics/sources` — Source usage breakdown (`GetSourceUsage` → `AnalyticsRepo.GetSourceUsageStats`)
  - Default `days=30`, max 365

