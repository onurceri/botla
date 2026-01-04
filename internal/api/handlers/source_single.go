package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/processing"
)

// GetSourceStatusOrDelete handles GET/DELETE for individual sources
func (h *SourcesHandlers) GetSourceStatusOrDelete(w http.ResponseWriter, r *http.Request) {
	s, _, _, ok := getSourceContext(w, r, h.SourceRepo, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getSourceStatus(w, r, s)
	case http.MethodDelete:
		h.deleteSource(w, r, s)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// getSourceStatus returns source status with ETag support
func (h *SourcesHandlers) getSourceStatus(w http.ResponseWriter, r *http.Request, s *models.DataSource) {
	// Compute ETag from status + processed_at + chunk_count
	etag := s.Status
	if s.ProcessedAt != nil {
		etag += "-" + s.ProcessedAt.UTC().Format(time.RFC3339Nano)
	}
	etag += "-" + strconv.Itoa(s.ChunkCount)

	inm := r.Header.Get("If-None-Match")
	if inm != "" && inm == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "private, must-revalidate")
	api.WriteJSON(w, http.StatusOK, s)
}

// deleteSource handles source deletion
func (h *SourcesHandlers) deleteSource(w http.ResponseWriter, r *http.Request, s *models.DataSource) {
	// Best-effort: delete associated vectors then remove source record
	if err := processing.DeleteSourceVectors(r.Context(), h.QdrantClient, s.ID); err != nil {
		h.logWarn("vector_delete_error", map[string]any{"source_id": s.ID, "error": err.Error()})
	}

	// Also delete from storage if it's a file
	if s.FilePath != nil && h.Storage != nil {
		_ = h.Storage.DeleteFile(r.Context(), *s.FilePath)
	}

	if err := h.SourceRepo.SoftDelete(r.Context(), s.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Re-aggregate suggestions after source deletion (use background context)
	// HI-002: Wrap in panic recovery to prevent server crash
	go func() {
		defer func() {
			if r := recover(); r != nil && h.Log != nil {
				h.Log.Error("re_aggregate_panic", map[string]any{"panic": r, "chatbot_id": s.ChatbotID})
			}
		}()
		processing.ReAggregateSuggestionsForChatbot(context.Background(), h.ChatbotRepo, s.ChatbotID, h.Log)
	}()

	w.WriteHeader(http.StatusNoContent)
}
