package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// AdminChatbotHandlers handles admin chatbot management endpoints
type AdminChatbotHandlers struct {
	DB           *sql.DB
	AdminService *services.AdminService
	RagService   *services.RAGService
	Queue        *services.Queue
}

// NewAdminChatbotHandlers creates a new AdminChatbotHandlers instance
func NewAdminChatbotHandlers(database *sql.DB, adminSvc *services.AdminService, ragSvc *services.RAGService, queue *services.Queue) *AdminChatbotHandlers {
	return &AdminChatbotHandlers{
		DB:           database,
		AdminService: adminSvc,
		RagService:   ragSvc,
		Queue:        queue,
	}
}

// ListChatbots returns a paginated list of all chatbots on the platform
func (h *AdminChatbotHandlers) ListChatbots(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	filter := db.ChatbotFilter{}
	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}
	if orgID := r.URL.Query().Get("organization_id"); orgID != "" {
		filter.OrganizationID = &orgID
	}
	if ownerID := r.URL.Query().Get("owner_id"); ownerID != "" {
		filter.OwnerID = &ownerID
	}

	chatbots, total, err := db.AdminListChatbots(r.Context(), h.DB, filter, limit, offset)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to list chatbots", api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"chatbots": chatbots,
		"total":    total,
	})
}

// GetChatbot returns details for a single chatbot
func (h *AdminChatbotHandlers) GetChatbot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing chatbot ID", api.ErrCodeBadRequest)
		return
	}

	chatbot, err := db.AdminGetChatbot(r.Context(), h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteError(w, http.StatusNotFound, "Chatbot not found", api.ErrCodeNotFound)
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "Failed to get chatbot", api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, chatbot)
}

// ForceRefreshChatbot resets all sources for a chatbot and triggers reprocessing
func (h *AdminChatbotHandlers) ForceRefreshChatbot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing chatbot ID", api.ErrCodeBadRequest)
		return
	}

	// Check if chatbot exists
	chatbot, err := db.AdminGetChatbot(r.Context(), h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteError(w, http.StatusNotFound, "Chatbot not found", api.ErrCodeNotFound)
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "Failed to get chatbot", api.ErrCodeInternalError)
		return
	}

	// Delete vectors from Qdrant if available
	if h.RagService != nil {
		_ = h.RagService.DeleteBotVectors(r.Context(), id)
	}

	// Reset chunk counts
	if err := db.AdminDeleteChatbotVectors(r.Context(), h.DB, id); err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to reset vectors", api.ErrCodeInternalError)
		return
	}

	// Reset all sources to pending
	count, err := db.AdminResetChatbotSources(r.Context(), h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to reset sources", api.ErrCodeInternalError)
		return
	}

	// Get pending source IDs and queue them
	sourceIDs, err := db.AdminGetChatbotSourceIDs(r.Context(), h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to get source IDs", api.ErrCodeInternalError)
		return
	}

	// Queue sources for processing
	queuedCount := 0
	if h.Queue != nil {
		for _, sourceID := range sourceIDs {
			if err := h.Queue.Enqueue(sourceID); err == nil {
				queuedCount++
			}
		}
	}

	// Log the action
	adminID, _ := middleware.UserIDFromContext(r.Context())
	_ = h.AdminService.LogAction(r.Context(), adminID, "force_refresh_chatbot", "chatbot", &id, map[string]any{
		"chatbot_name":   chatbot.Name,
		"sources_reset":  count,
		"sources_queued": queuedCount,
	}, r)

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"status":         "refreshing",
		"sources_reset":  count,
		"sources_queued": queuedCount,
	})
}


