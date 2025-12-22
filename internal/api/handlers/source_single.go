package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// GetSourceStatusOrDelete handles GET/DELETE for individual sources
func (h *SourcesHandlers) GetSourceStatusOrDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sourceID, ok := parseSourceIDFromPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s, err := db.GetSourceByID(r.Context(), h.DB, sourceID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if s == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	c, err := db.GetChatbotByID(r.Context(), h.DB, s.ChatbotID)
	if err != nil {
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(s); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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

	if err := db.SoftDeleteSource(r.Context(), h.DB, s.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Re-aggregate suggestions after source deletion (use background context)
	go processing.ReAggregateSuggestionsForChatbot(context.Background(), h.DB, s.ChatbotID, h.Log)

	w.WriteHeader(http.StatusNoContent)
}
