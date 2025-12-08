package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// ChatbotSources routes GET/POST requests for chatbot sources
func (h *SourcesHandlers) ChatbotSources(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	chatbotID, ok := parseChatbotIDFromPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if chatbotID == "new" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := db.GetChatbotByID(r.Context(), h.DB, chatbotID)
	if err != nil {
		h.logError("chatbot_fetch_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID, "path": r.URL.Path})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if c.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listSources(w, r, chatbotID)
	case http.MethodPost:
		h.createSource(w, r, chatbotID, userID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// listSources handles GET request to list all sources for a chatbot
func (h *SourcesHandlers) listSources(w http.ResponseWriter, r *http.Request, chatbotID string) {
	items, err := db.ListSourcesByChatbotID(r.Context(), h.DB, chatbotID)
	if err != nil {
		h.logError("sources_list_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID, "path": r.URL.Path})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
