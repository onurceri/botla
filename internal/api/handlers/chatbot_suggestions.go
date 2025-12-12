package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// SuggestionsHandlers handles suggestion-related endpoints.
type SuggestionsHandlers struct {
	DB  *sql.DB
	Log *logger.Logger
}

// RegenerateSuggestions handles POST /api/v1/chatbots/{id}/suggestions/regenerate
// It re-extracts and aggregates suggestions from all sources.
func (h *SuggestionsHandlers) RegenerateSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	chatbotID := extractChatbotIDFromPath(r.URL.Path)
	if chatbotID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify ownership
	bot, err := db.GetChatbotByID(r.Context(), h.DB, chatbotID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if bot.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Trigger re-aggregation (use background context since HTTP response will complete)
	go processing.ReAggregateSuggestionsForChatbot(context.Background(), h.DB, chatbotID, h.Log)

	w.WriteHeader(http.StatusAccepted)
}

// extractChatbotIDFromPath extracts chatbot ID from /api/v1/chatbots/{id}/suggestions/regenerate
func extractChatbotIDFromPath(path string) string {
	// Expected format: /api/v1/chatbots/{id}/suggestions/regenerate
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "chatbots" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
