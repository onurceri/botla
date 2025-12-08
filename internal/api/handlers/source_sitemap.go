package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// DiscoverSitemap handles POST /api/v1/chatbots/:id/sitemap/discover
// Parses a sitemap URL and returns all discovered URLs, optionally filtered by path filters
func (h *SourcesHandlers) DiscoverSitemap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Extract chatbot ID from path: /api/v1/chatbots/:id/sitemap/discover
	chatbotID, ok := parseSitemapDiscoverPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Verify chatbot ownership
	chatbot, err := db.GetChatbotByID(r.Context(), h.DB, chatbotID)
	if err != nil {
		h.logError("chatbot_fetch_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if chatbot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if chatbot.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		SitemapURL string `json:"sitemap_url"`
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate sitemap URL
	if err = scraper.ValidateSitemapURL(req.SitemapURL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse the sitemap
	parser := scraper.DefaultSitemapParser()
	result, err := parser.ParseSitemap(r.Context(), req.SitemapURL)
	if err != nil {
		h.logError("sitemap_parse_error", map[string]any{"error": err.Error(), "url": req.SitemapURL})
		http.Error(w, "Failed to parse sitemap: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Apply path filters if configured
	var filteredURLs []scraper.SitemapURL
	if len(chatbot.IncludePaths) > 0 || len(chatbot.ExcludePaths) > 0 {
		pathFilter, filterErr := scraper.NewPathFilter(chatbot.IncludePaths, chatbot.ExcludePaths)
		if filterErr == nil {
			filteredURLs = scraper.FilterURLsByPath(result.URLs, pathFilter)
		} else {
			filteredURLs = result.URLs
		}
	} else {
		filteredURLs = result.URLs
	}

	// Deduplicate
	filteredURLs = scraper.DeduplicateURLs(filteredURLs)

	// Build response
	response := struct {
		URLs       []scraper.SitemapURL `json:"urls"`
		TotalCount int                  `json:"total_count"`
	}{
		URLs:       filteredURLs,
		TotalCount: len(filteredURLs),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
