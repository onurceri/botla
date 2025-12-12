package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// BulkCreateSources handles POST /api/v1/chatbots/:id/sources/bulk
// Creates multiple URL sources at once
func (h *SourcesHandlers) BulkCreateSources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Extract chatbot ID from path: /api/v1/chatbots/:id/sources/bulk
	chatbotID, ok := parseBulkSourcesPath(r.URL.Path)
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
	base := api.BaseLang(chatbot.LanguageCode)
	cfg := api.ConfigFromBase(base)

	// Get plan for quota checks
	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil || plan == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req struct {
		URLs []string `json:"urls"`
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteLocalizedError(w, http.StatusBadRequest, api.ErrInvalidRequestBody, cfg)
		return
	}

	if len(req.URLs) == 0 {
		api.WriteLocalizedError(w, http.StatusBadRequest, api.ErrNoURLsProvided, cfg)
		return
	}

	// Check current URL count and limits
	currentCount, err := db.CountSourcesByType(r.Context(), h.DB, chatbotID, "url")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	limit := plan.Config.Scraping.MaxURLsPerBot
	if limit <= 0 {
		limit = 5 // Safe fallback
	}

	// Calculate how many URLs we can add
	available := limit - currentCount
	if available <= 0 {
		api.WriteLocalizedError(w, http.StatusForbidden, api.ErrURLLimitReached, cfg)
		return
	}

	// Check monthly ingestion quota
	usedSources, _, _ := db.GetMonthlyIngestionUsage(r.Context(), h.DB, userID, time.Now())
	maxIngest := plan.Config.MaxMonthlyIngestions
	if maxIngest <= 0 {
		maxIngest = 50
	}
	monthlyAvailable := maxIngest - usedSources
	if monthlyAvailable <= 0 {
		api.WriteLocalizedError(w, http.StatusPaymentRequired, api.ErrMonthlyIngestionExceeded, cfg)
		return
	}

	// Limit URLs to what's available
	urlsToProcess := req.URLs
	if len(urlsToProcess) > available {
		urlsToProcess = urlsToProcess[:available]
	}
	if len(urlsToProcess) > monthlyAvailable {
		urlsToProcess = urlsToProcess[:monthlyAvailable]
	}

	// Process each URL
	var createdCount, skippedCount int
	var errors []string

	for _, url := range urlsToProcess {
		url = strings.TrimSpace(url)
		if url == "" {
			skippedCount++
			continue
		}

		// Check for duplicates
		exists, _ := db.SourceExists(r.Context(), h.DB, chatbotID, url)
		if exists {
			skippedCount++
			continue
		}

		// Create the source
		ds := models.DataSource{
			ChatbotID:  chatbotID,
			SourceType: "url",
			Status:     "pending",
			SourceURL:  &url,
		}

		newID, createErr := db.CreateDataSource(r.Context(), h.DB, &ds)
		if createErr != nil {
			errors = append(errors, "Failed to create source for: "+url)
			continue
		}

		// Enqueue for processing
		if h.Queue != nil {
			h.Queue.Enqueue(newID)
		}
		createdCount++
	}

	// Build response
	response := struct {
		CreatedCount int      `json:"created_count"`
		SkippedCount int      `json:"skipped_count"`
		Errors       []string `json:"errors"`
	}{
		CreatedCount: createdCount,
		SkippedCount: skippedCount,
		Errors:       errors,
	}

	w.Header().Set("Content-Type", "application/json")
	if createdCount > 0 {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(response)
}
