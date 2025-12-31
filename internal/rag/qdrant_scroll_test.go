package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScrollChunks_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/collections/embeddings/points/scroll" {
			w.Header().Set("Content-Type", "application/json")

			// Verify request body if needed, but for now just mock response

			items := []SearchResult{
				{ID: "1", Score: 0.0, Payload: EmbeddingPayload{SourceID: "src"}},
				{ID: "2", Score: 0.0, Payload: EmbeddingPayload{SourceID: "src"}},
			}
			nextOffset := "next-uuid"

			resp := map[string]any{
				"status": "ok",
				"result": map[string]any{
					"points":           items,
					"next_page_offset": nextOffset,
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)

	c, _ := NewQdrantClient(&QdrantConfig{URL: srv.URL})

	// Assuming ScrollChunks signature: (ctx, sourceID, limit, offset) -> (points, nextOffset, error)
	points, next, err := c.ScrollChunks(context.Background(), "src", 10, nil)

	if err != nil {
		t.Fatalf("ScrollChunks returned error: %v", err)
	}

	if len(points) != 2 {
		t.Errorf("Expected 2 points, got %d", len(points))
	}

	if next == nil || *next != "next-uuid" {
		t.Errorf("Expected next offset 'next-uuid', got %v", next)
	}
}
