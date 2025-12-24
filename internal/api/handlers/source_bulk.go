package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

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

	chatbot, chatbotID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
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
	monthlyAvailable := h.getAvailableIngestionCount(r, userID, plan)
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

		_, createErr := h.persistAndEnqueueInternal(r, &ds)
		if createErr != nil {
			errors = append(errors, "Failed to create source for: "+url)
			continue
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

	if createdCount > 0 {
		api.WriteJSON(w, http.StatusCreated, response)
	} else {
		api.WriteJSON(w, http.StatusOK, response)
	}
}
