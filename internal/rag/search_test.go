package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
)

func TestSearchContext_Empty(t *testing.T) {
	mockVC := &MockVectorClient{}
	s, err := SearchContextTiered(context.Background(), mockVC, nil, "", 0, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Tier != TierLow {
		t.Fatalf("expected TierLow for empty input, got %s", s.Tier)
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

	// Use MediumThreshold 0.2 to filter out 0.1 score item
	cfg := &models.ThresholdConfig{MediumThreshold: 0.2, HighThreshold: 0.95}

	vc, _ := NewQdrantClientFromEnv()
	res, err := SearchContextTiered(context.Background(), vc, []float32{0.1}, "cb", 0, 0, cfg)
	if err != nil {
		t.Fatalf("search err: %v", err)
	}
	if res.ContextText == "" {
		t.Fatalf("empty body")
	}
	if len(res.Chunks) == 0 {
		t.Fatalf("no used metas")
	}
}

func TestSearchContext_MissingQdrant(t *testing.T) {
	t.Setenv("QDRANT_URL", "")
	// should handle missing qdrant gracefully
	vc, err := NewQdrantClientFromEnv()
	if err == nil {
		t.Fatalf("expected error from NewQdrantClientFromEnv")
	}
	_, err = SearchContextTiered(context.Background(), vc, []float32{0.1}, "cb", 0, 0, nil)
	if err == nil {
		t.Fatalf("expected error when qdrant client is nil")
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

	cfg := &models.ThresholdConfig{MediumThreshold: 0.2, HighThreshold: 0.8}
	vc, _ := NewQdrantClientFromEnv()
	res, err := SearchContextTiered(context.Background(), vc, []float32{0.1}, "cb", 0, 0, cfg)
	if err != nil {
		t.Fatalf("search err: %v", err)
	}
	if res.ContextText != "" {
		t.Fatalf("expected empty body when all below threshold")
	}
	if res.Tier != TierLow {
		t.Fatalf("expected TierLow, got %s", res.Tier)
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
	cfg := &models.ThresholdConfig{MediumThreshold: 0.0, HighThreshold: 0.5}
	vc, _ := NewQdrantClientFromEnv()
	res, err := SearchContextTiered(context.Background(), vc, []float32{0.1}, "cb", 5, 1000, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Chunks) != 2 {
		t.Errorf("expected 2 items, got %d", len(res.Chunks))
	}
	if !strings.Contains(res.ContextText, "ZeroScore") {
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
	cfg := &models.ThresholdConfig{MediumThreshold: 1.0, HighThreshold: 1.0}
	vc, _ := NewQdrantClientFromEnv()
	res, err := SearchContextTiered(context.Background(), vc, []float32{0.1}, "cb", 5, 1000, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Existing behavior: even if filtered, raw hits are returned in AllChunks.
	// But body should be empty.
	if len(res.AllChunks) != 1 {
		t.Errorf("expected 1 raw item in AllChunks, got %d", len(res.AllChunks))
	}
	if res.ContextText != "" {
		t.Errorf("expected empty body, got %q", res.ContextText)
	}
	if res.Tier != TierLow {
		t.Errorf("expected TierLow, got %s", res.Tier)
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

	cfg := &models.ThresholdConfig{MediumThreshold: 0.5, HighThreshold: 0.8}
	vc, _ := NewQdrantClientFromEnv()
	res, err := SearchContextTiered(context.Background(), vc, []float32{0.1}, "cb", 5, 1000, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The implementation sorts by score descending.
	// 0.9 (First) -> 0.8 (Second)
	// Output should be First + "\n---\n" + Second
	if !strings.Contains(res.ContextText, "First\n---\nSecond") {
		t.Errorf("expected separator between chunks. Got: %q", res.ContextText)
	}
}
