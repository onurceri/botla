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

// AdminSourceHandlers handles admin data source management endpoints
type AdminSourceHandlers struct {
	DB           *sql.DB
	AdminService *services.AdminService
	RagService   *services.RAGService
	Queue        *services.Queue
}

// NewAdminSourceHandlers creates a new AdminSourceHandlers instance
func NewAdminSourceHandlers(database *sql.DB, adminSvc *services.AdminService, ragSvc *services.RAGService, queue *services.Queue) *AdminSourceHandlers {
	return &AdminSourceHandlers{
		DB:           database,
		AdminService: adminSvc,
		RagService:   ragSvc,
		Queue:        queue,
	}
}

// ListSources returns a paginated list of all data sources on the platform
func (h *AdminSourceHandlers) ListSources(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	filter := db.SourceFilter{}
	if chatbotID := r.URL.Query().Get("chatbot_id"); chatbotID != "" {
		filter.ChatbotID = &chatbotID
	}
	if sourceType := r.URL.Query().Get("source_type"); sourceType != "" {
		filter.SourceType = &sourceType
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	if ownerID := r.URL.Query().Get("owner_id"); ownerID != "" {
		filter.OwnerID = &ownerID
	}

	sources, total, err := db.AdminListSources(r.Context(), h.DB, filter, limit, offset)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"sources": sources,
		"total":   total,
	})
}

// GetSource returns details for a single data source
func (h *AdminSourceHandlers) GetSource(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	source, err := db.AdminGetSourceByID(r.Context(), h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
			return
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, source)
}

// GetSourceStats returns aggregated statistics for data sources
func (h *AdminSourceHandlers) GetSourceStats(w http.ResponseWriter, r *http.Request) {
	stats, err := db.AdminGetSourceStats(r.Context(), h.DB)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, stats)
}

// ReprocessSource resets a source to pending status and queues it for reprocessing
func (h *AdminSourceHandlers) ReprocessSource(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	// Get source details for logging
	source, err := db.AdminGetSourceByID(r.Context(), h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
			return
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Delete existing vectors from Qdrant if available
	if h.RagService != nil {
		_ = h.RagService.DeleteSourceVectors(r.Context(), id)
	}

	// Reset source to pending
	if err := db.AdminReprocessSource(r.Context(), h.DB, id); err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Queue for processing
	queued := false
	if h.Queue != nil {
		if err := h.Queue.Enqueue(id); err == nil {
			queued = true
		}
	}

	// Log the action
	adminID, _ := middleware.UserIDFromContext(r.Context())
	_ = h.AdminService.LogAction(r.Context(), adminID, "reprocess_source", "source", &id, map[string]any{
		"source_type":  source.SourceType,
		"chatbot_id":   source.ChatbotID,
		"chatbot_name": source.ChatbotName,
		"queued":       queued,
	}, r)

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "reprocessing",
		"queued": queued,
	})
}

// GetSourceTypes returns available source types
func (h *AdminSourceHandlers) GetSourceTypes(w http.ResponseWriter, r *http.Request) {
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"types":    db.AdminListSourceTypes(),
		"statuses": db.AdminListSourceStatuses(),
	})
}
