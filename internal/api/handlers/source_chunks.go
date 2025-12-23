package handlers

import (
	"encoding/json"
	"net/http"
)

// GetSourceChunks retrieves the chunks for a specific source.
func (h *SourcesHandlers) GetSourceChunks(w http.ResponseWriter, r *http.Request) {
	_, _, sourceID, ok := getSourceContext(w, r, h.DB, h.WorkspaceService, h.OrgService, "/chunks")
	if !ok {
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
