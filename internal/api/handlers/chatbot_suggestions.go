package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/logger"
)

// SuggestionsHandlers handles suggestion-related endpoints.
type SuggestionsHandlers struct {
	DB               *sql.DB
	Log              *logger.Logger
	WorkspaceService *services.WorkspaceService
	OrgService       *services.OrganizationService
}

// RegenerateSuggestions handles POST /api/v1/chatbots/{id}/suggestions/regenerate
// It re-extracts and aggregates suggestions from all sources.
func (h *SuggestionsHandlers) RegenerateSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, chatbotID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Trigger re-aggregation (use background context since HTTP response will complete)
	go processing.ReAggregateSuggestionsForChatbot(context.Background(), h.DB, chatbotID, h.Log)

	w.WriteHeader(http.StatusAccepted)
}
