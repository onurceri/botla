package router

import (
	"net/http"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func chatbotsDispatchHandler(secret string, ch *handlers.ChatbotHandlers, sh *handlers.SourcesHandlers, chh *handlers.ChatHandlers, puh *handlers.PendingURLsHandlers, acth *handlers.ActionHandlers, hoh *handlers.HandoffHandlers, anh *handlers.AnalyticsHandlers, sugh *handlers.SuggestionsHandlers) http.Handler {
	mux := http.NewServeMux()

	// Pending URLs
	mux.HandleFunc("POST /api/v1/chatbots/{id}/pending-urls/approve", puh.ApprovePendingURLs)
	mux.HandleFunc("POST /api/v1/chatbots/{id}/pending-urls/reject", puh.RejectPendingURLs)
	mux.HandleFunc("DELETE /api/v1/chatbots/{id}/pending-urls/clear", puh.ClearPendingURLs)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/pending-urls", puh.ListPendingURLs)

	// Chat
	mux.HandleFunc("POST /api/v1/chatbots/{id}/chat", chh.Chat)

	// Actions (Legacy dispatch)
	// We keep the manual dispatch for actions because it handles sub-paths dynamically
	mux.HandleFunc("/api/v1/chatbots/{id}/actions/", acth.Dispatch)
	mux.HandleFunc("/api/v1/chatbots/{id}/actions", acth.Dispatch)

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

	// Sitemap
	mux.HandleFunc("POST /api/v1/chatbots/{id}/sitemap/discover", sh.DiscoverSitemap)

	// Sources
	mux.HandleFunc("POST /api/v1/chatbots/{id}/sources/bulk", sh.BulkCreateSources)
	mux.HandleFunc("GET /api/v1/chatbots/{id}/sources", sh.ChatbotSources)

	// Chatbot management (Fallback)
	mux.HandleFunc("/api/v1/chatbots/{id}", ch.ByID)

	return middleware.AuthMiddleware(secret)(mux)
}
