> Extracted from: chatbot-management.md — Section 3

# 3. API Endpoints Table

| Method | Path | Handler | Success Codes | Error Codes |
|--------|------|---------|---------------|-------------|
| `GET` | `/api/v1/chatbots` | `ChatbotHandlers.ListOrCreate` | `200` | `401`, `403`, `500` |
| `POST` | `/api/v1/chatbots` | `ChatbotHandlers.ListOrCreate` | `201` | `400`, `401`, `403`, `500` |
| `GET` | `/api/v1/chatbots/{id}` | `ChatbotHandlers.ByID` | `200` | `400`, `401`, `403`, `404`, `500` |
| `PUT` | `/api/v1/chatbots/{id}` | `ChatbotHandlers.ByID` | `200` | `400`, `401`, `403`, `404`, `500` |
| `DELETE` | `/api/v1/chatbots/{id}` | `ChatbotHandlers.ByID` | `204` | `401`, `403`, `404`, `500` |
| `PUT` | `/api/v1/chatbots/{id}/basic-info` | `ChatbotHandlers.UpdateBasicInfo` | `200` | `400`, `401`, `403`, `404`, `500` |
| `PUT` | `/api/v1/chatbots/{id}/appearance` | `ChatbotHandlers.UpdateAppearance` | `200` | `400`, `401`, `403`, `404`, `500` |
| `PUT` | `/api/v1/chatbots/{id}/model` | `ChatbotHandlers.UpdateModelSettings` | `200` | `400`, `401`, `403`, `404`, `500` |
| `PUT` | `/api/v1/chatbots/{id}/security` | `ChatbotHandlers.UpdateSecuritySettings` | `200` | `400`, `401`, `403`, `404`, `500` |
| `PUT` | `/api/v1/chatbots/{id}/guardrails` | `ChatbotHandlers.UpdateGuardrails` | `200` | `400`, `401`, `403`, `404`, `500` |
| `PUT` | `/api/v1/chatbots/{id}/handoff` | `ChatbotHandlers.UpdateHandoff` | `200` | `400`, `401`, `403`, `404`, `500` |
| `PUT` | `/api/v1/chatbots/{id}/refresh` | `ChatbotHandlers.UpdateRefresh` | `200` | `400`, `401`, `403`, `404`, `500` |
| `PUT` | `/api/v1/chatbots/{id}/scraping` | `ChatbotHandlers.UpdateScrapingConfig` | `200` | `400`, `401`, `403`, `404`, `500` |
| `POST` | `/api/v1/chatbots/{id}/suggestions/regenerate` | `SuggestionsHandlers.RegenerateSuggestions` | `202` | `401`, `403`, `404`, `500` |
| `GET` | `/api/v1/chatbots/{id}/suggestions/status` | `SuggestionsHandlers.GetSuggestionJobStatus` | `200` | `401`, `403`, `404`, `500` |
| `POST` | `/api/v1/chatbots/{id}/sitemap/discover` | `SourcesHandlers.DiscoverSitemap` | `202` | `401`, `403`, `404`, `500` |
