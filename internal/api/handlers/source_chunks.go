package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/pkg/middleware"
)

// GetSourceChunks retrieves the chunks for a specific source.
func (h *SourcesHandlers) GetSourceChunks(w http.ResponseWriter, r *http.Request) {
	// 1. Parse SourceID from URL
	// URL format: /api/v1/sources/{sourceID}/chunks
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 { // expecting at least "", "api", "v1", "sources", "ID", "chunks"
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	sourceID := parts[4]

	// 2. Auth & Ownership Check
	userID, ok := r.Context().Value(middleware.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if source belongs to user (via chatbot)
	var exists bool
	err := h.DB.QueryRowContext(r.Context(), `
        SELECT EXISTS(
            SELECT 1 FROM data_sources s 
            JOIN chatbots c ON s.chatbot_id = c.id 
            WHERE s.id = $1 AND c.user_id = $2
        )`, sourceID, userID).Scan(&exists)

	if err != nil {
		h.Log.Error("db_error", map[string]any{"error": err.Error()})
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "source not found or access denied", http.StatusNotFound)
		return
	}

	// 3. Fetch Chunks from Qdrant
	offsetParam := r.URL.Query().Get("offset")
	var offset interface{}
	if offsetParam != "" {
		offset = offsetParam
	}

	// Limit
	limit := 20

	points, nextOffset, err := h.QdrantClient.ScrollChunks(r.Context(), sourceID, limit, offset)
	if err != nil {
		h.Log.Error("qdrant_scroll_error", map[string]any{"error": err.Error(), "source_id": sourceID})
		http.Error(w, "failed to retrieve chunks", http.StatusInternalServerError)
		return
	}

	// 4. Response
	resp := map[string]any{
		"chunks":      points,
		"next_cursor": nextOffset,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Log.Error("json_encode_error", map[string]any{"error": err.Error()})
	}
}
