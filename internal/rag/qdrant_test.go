package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewQdrantClient_ValidConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := &QdrantConfig{
		URL:    srv.URL,
		APIKey: "test-key",
	}

	c, err := NewQdrantClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected client to be created")
	}
	if c.baseURL != srv.URL {
		t.Errorf("expected baseURL %s, got %s", srv.URL, c.baseURL)
	}
	if c.apiKey != "test-key" {
		t.Errorf("expected apiKey test-key, got %s", c.apiKey)
	}
	if c.http == nil {
		t.Error("expected http client to be set")
	}
}

func TestNewQdrantClient_NilConfig(t *testing.T) {
	_, err := NewQdrantClient(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestNewQdrantClient_EmptyURL(t *testing.T) {
	cfg := &QdrantConfig{
		URL:    "",
		APIKey: "test-key",
	}
	_, err := NewQdrantClient(cfg)
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestNewQdrantClient_EmptyAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := &QdrantConfig{
		URL:    srv.URL,
		APIKey: "",
	}

	c, err := NewQdrantClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.apiKey != "" {
		t.Error("expected empty API key to be allowed")
	}
}

func TestNewQdrantClient_DefaultTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := &QdrantConfig{
		URL: srv.URL,
	}

	c, err := NewQdrantClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.http.Timeout != 15*time.Second {
		t.Errorf("expected default timeout 15s, got %v", c.http.Timeout)
	}
}

func TestNewQdrantClient_CustomTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := &QdrantConfig{
		URL:     srv.URL,
		Timeout: 30 * time.Second,
	}

	c, err := NewQdrantClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.http.Timeout != 30*time.Second {
		t.Errorf("expected custom timeout 30s, got %v", c.http.Timeout)
	}
}

func TestNewQdrantClientFromEnv(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()
	cfg := &QdrantConfig{URL: srv.URL}
	c, err := NewQdrantClient(cfg)
	if err != nil || c == nil {
		t.Fatalf("client err: %v", err)
	}
}

func TestNewQdrantClientFromEnv_MissingURL(t *testing.T) {
	cfg := &QdrantConfig{URL: ""}
	c, err := NewQdrantClient(cfg)
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
	cfg := &QdrantConfig{URL: srv.URL}
	c, _ := NewQdrantClient(cfg)
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
	cfg := &QdrantConfig{URL: srv.URL, APIKey: "key123"}
	c, _ := NewQdrantClient(cfg)
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
	cfg := &QdrantConfig{URL: srv.URL}
	c, _ := NewQdrantClient(cfg)
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
	cfg := &QdrantConfig{URL: srv.URL}
	c, _ := NewQdrantClient(cfg)
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
	cfg := &QdrantConfig{URL: srv.URL}
	c, _ := NewQdrantClient(cfg)
	if err := c.DeleteBySourceID(context.Background(), "src"); err != nil {
		t.Fatalf("delete err: %v", err)
	}
}
