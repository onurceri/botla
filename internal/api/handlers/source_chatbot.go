package handlers

import (
	"net/http"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/pkg/middleware"
)

// ChatbotSources routes GET/POST requests for chatbot sources
func (h *SourcesHandlers) ChatbotSources(w http.ResponseWriter, r *http.Request) {
	bot, chatbotID, ok := getChatbotContext(w, r, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	switch r.Method {
	case http.MethodGet:
		h.listSources(w, r, chatbotID)
	case http.MethodPost:
		h.createSource(w, r, bot, userID)
	default:
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// listSources handles GET request to list all sources for a chatbot
func (h *SourcesHandlers) listSources(w http.ResponseWriter, r *http.Request, chatbotID string) {
	items, err := h.SourceRepo.GetByChatbot(r.Context(), chatbotID)
	if err != nil {
		h.logError("sources_list_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID, "path": r.URL.Path})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	api.WriteJSON(w, http.StatusOK, items)
}
