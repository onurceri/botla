package openai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/ai"
)

func TestEmbedder_ImplementsInterface(t *testing.T) {
	var _ ai.Embedder = (*Embedder)(nil)
}

func TestNewEmbedder(t *testing.T) {
	embedder, err := NewEmbedder(Config{
		APIKey:  "test-key",
		BaseURL: "https://api.openai.com",
		Model:   "text-embedding-3-small",
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if embedder == nil {
		t.Fatal("expected embedder to be created")
	}
	if embedder.client == nil {
		t.Error("expected client to be set")
	}
	if embedder.model != "text-embedding-3-small" {
		t.Errorf("expected model to be 'text-embedding-3-small', got %s", embedder.model)
	}
}

func TestNewEmbedder_EmptyAPIKey(t *testing.T) {
	_, err := NewEmbedder(Config{BaseURL: "https://api.openai.com"}, nil)
	if err == nil {
		t.Error("expected error when API key is empty")
	}
}

func TestNewEmbedder_WithDefaults(t *testing.T) {
	embedder, err := NewEmbedder(Config{APIKey: "test-key"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if embedder.model != "text-embedding-3-small" {
		t.Errorf("expected default model 'text-embedding-3-small', got %s", embedder.model)
	}
}

func TestEmbed_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/embeddings" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"data": [
					{
						"embedding": [0.1, 0.2, 0.3]
					}
				],
				"usage": {
					"prompt_tokens": 10,
					"total_tokens": 10
				}
			}`))
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	embedder, err := NewEmbedder(Config{APIKey: "test-key", BaseURL: server.URL, Model: "text-embedding-3-small"}, server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	result, err := embedder.Embed(ctx, "test text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 dimensions, got %d", len(result))
	}
	if result[0] != 0.1 || result[1] != 0.2 || result[2] != 0.3 {
		t.Errorf("unexpected embedding values: %v", result)
	}
}

func TestEmbed_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	embedder, err := NewEmbedder(Config{APIKey: "test-key", BaseURL: server.URL, Model: "text-embedding-3-small"}, server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	_, err = embedder.Embed(ctx, "test text")
	if err == nil {
		t.Error("expected error for unauthorized request")
	}
}

func TestEmbedBatch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/embeddings" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"data": [
					{
						"embedding": [0.1, 0.2, 0.3]
					},
					{
						"embedding": [0.4, 0.5, 0.6]
					}
				],
				"usage": {
					"prompt_tokens": 20,
					"total_tokens": 20
				}
			}`))
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	embedder, err := NewEmbedder(Config{APIKey: "test-key", BaseURL: server.URL, Model: "text-embedding-3-small"}, server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	results, err := embedder.EmbedBatch(ctx, []string{"text 1", "text 2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if len(results[0]) != 3 || len(results[1]) != 3 {
		t.Error("unexpected embedding dimensions")
	}
	if results[0][0] != 0.1 || results[1][0] != 0.4 {
		t.Errorf("unexpected embedding values: %v", results)
	}
}

func TestEmbedBatch_EmptyInput(t *testing.T) {
	embedder, err := NewEmbedder(Config{APIKey: "test-key", BaseURL: "https://api.openai.com", Model: "text-embedding-3-small"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	results, err := embedder.EmbedBatch(ctx, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty input, got %v", results)
	}
}

func TestDimension(t *testing.T) {
	tests := []struct {
		model    string
		expected int
	}{
		{"text-embedding-3-small", 1536},
		{"text-embedding-3-large", 3072},
		{"custom-model", 1536}, // default
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			embedder, err := NewEmbedder(Config{APIKey: "test-key", BaseURL: "https://api.openai.com", Model: tt.model}, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dim := embedder.Dimension()
			if dim != tt.expected {
				t.Errorf("expected dimension %d, got %d", tt.expected, dim)
			}
		})
	}
}

func TestEmbed_NilClient(t *testing.T) {
	var embedder *Embedder
	ctx := context.Background()

	_, err := embedder.Embed(ctx, "test")
	if err == nil {
		t.Error("expected error for nil embedder")
	}
}

func TestEmbedBatch_NilClient(t *testing.T) {
	var embedder *Embedder
	ctx := context.Background()

	_, err := embedder.EmbedBatch(ctx, []string{"test"})
	if err == nil {
		t.Error("expected error for nil embedder")
	}
}
