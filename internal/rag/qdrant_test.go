package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewQdrantClientFromEnv(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)
	c, err := NewQdrantClientFromEnv()
	if err != nil || c == nil {
		t.Fatalf("client err: %v", err)
	}
}

func TestNewQdrantClientFromEnv_MissingURL(t *testing.T) {
	t.Setenv("QDRANT_URL", "")
	c, err := NewQdrantClientFromEnv()
	if err == nil || c != nil {
		t.Fatalf("expected error for missing url")
	}
}

func TestEnsureEmbeddingsCollection_Create(t *testing.T) {
	created := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/collections/embeddings":
			http.NotFound(w, r)
		case r.Method == http.MethodPut && r.URL.Path == "/collections/embeddings":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			created = true
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)
	c, _ := NewQdrantClientFromEnv()
	if err := c.EnsureEmbeddingsCollection(context.Background()); err != nil {
		t.Fatalf("ensure err: %v", err)
	}
	if !created {
		t.Fatalf("collection not created")
	}
}

func TestEnsureEmbeddingsCollection_AuthHeader(t *testing.T) {
	got := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/collections/embeddings" {
			got = r.Header.Get("api-key")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)
	t.Setenv("QDRANT_API_KEY", "key123")
	c, _ := NewQdrantClientFromEnv()
	if err := c.EnsureEmbeddingsCollection(context.Background()); err != nil {
		t.Fatalf("ensure err: %v", err)
	}
	if got != "key123" {
		t.Fatalf("api-key header not set")
	}
}

func TestUpsertEmbedding_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/collections/embeddings/points" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)
	c, _ := NewQdrantClientFromEnv()
	err := c.UpsertEmbedding(context.Background(), "id", []float32{0.1, 0.2}, EmbeddingPayload{})
	if err != nil {
		t.Fatalf("upsert err: %v", err)
	}
}

func TestSearchSimilar_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/collections/embeddings/points/search" {
			w.Header().Set("Content-Type", "application/json")
			// qdrantResponse with a list of items in result
			items := []SearchResult{{ID: "1", Score: 0.9, Payload: EmbeddingPayload{ChatbotID: "cb"}}}
			json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": items})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)
	c, _ := NewQdrantClientFromEnv()
	res, err := c.SearchSimilar(context.Background(), []float32{0.1}, "cb", 1)
	if err != nil || len(res) != 1 {
		t.Fatalf("search err: %v", err)
	}
}

func TestDeleteBySourceID_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/collections/embeddings/points/delete" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)
	c, _ := NewQdrantClientFromEnv()
	if err := c.DeleteBySourceID(context.Background(), "src"); err != nil {
		t.Fatalf("delete err: %v", err)
	}
}
