package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/pkg/middleware"
)

// BulkCreateSources handles POST /api/v1/chatbots/:id/sources/bulk
// Creates multiple URL sources at once
func (h *SourcesHandlers) BulkCreateSources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, chatbotID, ok := getChatbotContext(w, r, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())

	// Get plan for quota checks
	plan, err := h.PlanRepo.GetPlanWithLimits(r.Context(), userID)
	if err != nil || plan == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req struct {
		URLs []string `json:"urls"`
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidRequestBody)
		return
	}

	if len(req.URLs) == 0 {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrNoURLsProvided)
		return
	}

	// Check current URL count and limits
	currentCount, err := h.SourceRepo.CountByType(r.Context(), chatbotID, "url")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	limit := plan.Limits.ScrapingMaxURLsPerBot
	if limit <= 0 {
		limit = 5 // Safe fallback
	}

	// Calculate how many URLs we can add
	available := limit - currentCount
	if available <= 0 {
		api.WriteErrorCode(w, http.StatusForbidden, api.ErrURLLimitReached)
		return
	}

	// Check monthly ingestion quota
	monthlyAvailable := h.getAvailableIngestionCount(r, userID, plan)
	if monthlyAvailable <= 0 {
		api.WriteErrorCode(w, http.StatusPaymentRequired, api.ErrMonthlyIngestionExceeded)
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
		exists, _ := h.SourceRepo.Exists(r.Context(), chatbotID, url)
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
