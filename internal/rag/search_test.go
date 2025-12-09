package rag

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchContext_Empty(t *testing.T) {
	s, metas, err := SearchContext(nil, "", 0, 0, 0)
	if err != nil || s != "" || metas != nil {
		t.Fatalf("unexpected for empty input")
	}
}

func TestSearchContext_ThresholdAndMaxTokens(t *testing.T) {
	// qdrant responds with two items, one below threshold
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points/search" {
			items := []SearchResult{
				{ID: "1", Score: 0.9, Payload: EmbeddingPayload{OriginalText: "First", SourceID: "s1", SourceType: "file", ChunkIndex: 0}},
				{ID: "2", Score: 0.1, Payload: EmbeddingPayload{OriginalText: "Second", SourceID: "s2", SourceType: "file", ChunkIndex: 1}},
			}
			json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": items})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)
	t.Setenv("RAG_TOPK", "5")
	t.Setenv("RAG_MAX_CONTEXT_TOKENS", "5")
	body, used, err := SearchContext([]float32{0.1}, "cb", 0, 0, 0.2)
	if err != nil {
		t.Fatalf("search err: %v", err)
	}
	if body == "" {
		t.Fatalf("empty body")
	}
	if len(used) == 0 {
		t.Fatalf("no used metas")
	}
}

func TestSearchContext_MissingQdrant(t *testing.T) {
	t.Setenv("QDRANT_URL", "")
	// should handle missing qdrant gracefully if SearchContext is called
	// (though SearchContext checks err != nil from NewQdrantClient)
	_, _, err := SearchContext([]float32{0.1}, "cb", 0, 0, 0)
	if err == nil {
		t.Fatalf("expected error when qdrant url missing")
	}
}

func TestSearchContext_AllBelowThreshold(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points/search" {
			items := []SearchResult{
				{ID: "1", Score: 0.1, Payload: EmbeddingPayload{OriginalText: "First"}},
				{ID: "2", Score: 0.19, Payload: EmbeddingPayload{OriginalText: "Second"}},
			}
			json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": items})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)
	body, metas, err := SearchContext([]float32{0.1}, "cb", 0, 0, 0.2)
	if err != nil {
		t.Fatalf("search err: %v", err)
	}
	if body != "" {
		t.Fatalf("expected empty body when all below threshold")
	}
	if metas == nil || len(metas) != 2 {
		t.Fatalf("metas should include raw hits")
	}
}

// RAG-008: Score threshold at exactly 0.0 (permissive)
func TestSearchContext_ThresholdZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points/search" {
			items := []SearchResult{
				{ID: "1", Score: 0.0, Payload: EmbeddingPayload{OriginalText: "ZeroScore"}},
				{ID: "2", Score: 0.1, Payload: EmbeddingPayload{OriginalText: "PositiveScore"}},
			}
			json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": items})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)

	// Threshold 0.0 should include the 0.0 score item
	body, metas, err := SearchContext([]float32{0.1}, "cb", 5, 1000, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(metas) != 2 {
		t.Errorf("expected 2 items, got %d", len(metas))
	}
	if !strings.Contains(body, "ZeroScore") {
		t.Errorf("expected ZeroScore text in body")
	}
}

// RAG-009: Score threshold at 1.0 (restrictive)
func TestSearchContext_ThresholdOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points/search" {
			items := []SearchResult{
				{ID: "1", Score: 0.99, Payload: EmbeddingPayload{OriginalText: "AlmostPerfect"}},
			}
			json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": items})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)

	// Threshold 1.0 should exclude 0.99
	body, metas, err := SearchContext([]float32{0.1}, "cb", 5, 1000, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Existing behavior: even if filtered, raw hits are returned in metas.
	// But body should be empty.
	if len(metas) != 1 {
		t.Errorf("expected 1 raw item in metas, got %d", len(metas))
	}
	if body != "" {
		t.Errorf("expected empty body")
	}
}

// RAG-010: Context aggregation separator
func TestSearchContext_Separator(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points/search" {
			items := []SearchResult{
				{ID: "1", Score: 0.9, Payload: EmbeddingPayload{OriginalText: "First"}},
				{ID: "2", Score: 0.8, Payload: EmbeddingPayload{OriginalText: "Second"}},
			}
			json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": items})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)

	body, _, err := SearchContext([]float32{0.1}, "cb", 5, 1000, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The implementation sorts by score descending.
	// 0.9 (First) -> 0.8 (Second)
	// Output should be First + "\n---\n" + Second
	if !strings.Contains(body, "First\n---\nSecond") {
		t.Errorf("expected separator between chunks. Got: %q", body)
	}
}
