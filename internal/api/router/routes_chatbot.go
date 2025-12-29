package router

import (
	"net/http"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func ChatbotsDispatchHandler(secret string, ch *handlers.ChatbotHandlers, sh *handlers.SourcesHandlers, chh *handlers.ChatHandlers, puh *handlers.PendingURLsHandlers, acth *handlers.ActionHandlers, hoh *handlers.HandoffHandlers, anh *handlers.AnalyticsHandlers, sugh *handlers.SuggestionsHandlers) http.Handler {
	return middleware.AuthMiddleware(secret)(middleware.ExtractTenantContext()(ChatbotsRawHandler(ch, sh, chh, puh, acth, hoh, anh, sugh)))
}

func ChatbotsRawHandler(ch *handlers.ChatbotHandlers, sh *handlers.SourcesHandlers, chh *handlers.ChatHandlers, puh *handlers.PendingURLsHandlers, acth *handlers.ActionHandlers, hoh *handlers.HandoffHandlers, anh *handlers.AnalyticsHandlers, sugh *handlers.SuggestionsHandlers) http.Handler {
	mux := http.NewServeMux()

	// Pending URLs
	mux.HandleFunc("POST /api/v1/chatbots/{id}/pending-urls/approve", puh.ApprovePendingURLs)
	mux.HandleFunc("POST /api/v1/chatbots/{id}/pending-urls/reject", puh.RejectPendingURLs)
	mux.HandleFunc("DELETE /api/v1/chatbots/{id}/pending-urls/clear", puh.ClearPendingURLs)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/pending-urls", puh.ListPendingURLs)

	// Chat
	mux.HandleFunc("POST /api/v1/chatbots/{id}/chat", chh.Chat)

	// Actions
	mux.HandleFunc("GET /api/v1/chatbots/{id}/actions", acth.List)
	mux.HandleFunc("POST /api/v1/chatbots/{id}/actions", acth.Create)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/actions/logs", acth.GetLogs)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/actions/{actionId}", acth.Get)
	mux.HandleFunc("PUT /api/v1/chatbots/{id}/actions/{actionId}", acth.Update)
	mux.HandleFunc("DELETE /api/v1/chatbots/{id}/actions/{actionId}", acth.Delete)

	// Handoff Requests
	mux.HandleFunc("GET /api/v1/chatbots/{id}/handoff-requests", hoh.ListHandoffRequests)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/handoff-requests/{requestId}", hoh.GetHandoffRequestDetail)
	mux.HandleFunc("PATCH /api/v1/chatbots/{id}/handoff-requests/{requestId}", hoh.UpdateHandoffRequest)

	// Analytics
	mux.HandleFunc("GET /api/v1/chatbots/{id}/analytics/overview", anh.GetChatbotAnalyticsOverview)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/analytics/trends", anh.GetChatbotAnalyticsTrends)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/analytics/sources", anh.GetSourceUsage)

	// Suggestions
	mux.HandleFunc("POST /api/v1/chatbots/{id}/suggestions/regenerate", sugh.RegenerateSuggestions)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/suggestions/status", sugh.GetSuggestionJobStatus)

	// Sitemap
	mux.HandleFunc("POST /api/v1/chatbots/{id}/sitemap/discover", sh.DiscoverSitemap)

	// Sources
	mux.HandleFunc("POST /api/v1/chatbots/{id}/sources/bulk", sh.BulkCreateSources)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/sources", sh.ChatbotSources)
	mux.HandleFunc("POST /api/v1/chatbots/{id}/sources", sh.ChatbotSources)

	// Domain-specific updates
	mux.HandleFunc("PUT /api/v1/chatbots/{id}/basic-info", ch.UpdateBasicInfo)
	mux.HandleFunc("PUT /api/v1/chatbots/{id}/appearance", ch.UpdateAppearance)
	mux.HandleFunc("PUT /api/v1/chatbots/{id}/model", ch.UpdateModelSettings)
	mux.HandleFunc("PUT /api/v1/chatbots/{id}/security", ch.UpdateSecuritySettings)
	mux.HandleFunc("PUT /api/v1/chatbots/{id}/guardrails", ch.UpdateGuardrails)
	mux.HandleFunc("PUT /api/v1/chatbots/{id}/handoff", ch.UpdateHandoff)
	mux.HandleFunc("PUT /api/v1/chatbots/{id}/refresh", ch.UpdateRefresh)
	mux.HandleFunc("PUT /api/v1/chatbots/{id}/scraping", ch.UpdateScrapingConfig)

	// Chatbot management (Fallback)
	mux.HandleFunc("/api/v1/chatbots/{id}", ch.ByID)

	return mux
}
