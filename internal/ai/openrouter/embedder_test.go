package openrouter

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
	embedder := NewEmbedder("test-key", "https://openrouter.ai/api/v1", "text-embedding-3-small", nil)
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

func TestNewFromEnv_Success(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "test-key")
	t.Setenv("OPENROUTER_API_BASE", "https://openrouter.ai/api/v1")
	t.Setenv("OPENROUTER_EMBEDDING_MODEL", "text-embedding-3-large")

	embedder, err := NewFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if embedder.client == nil {
		t.Error("expected client to be set")
	}
	if embedder.model != "text-embedding-3-large" {
		t.Errorf("expected model to be 'text-embedding-3-large', got %s", embedder.model)
	}
}

func TestNewFromEnv_FallbackToOpenAI(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "fallback-key")

	embedder, err := NewFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if embedder.client == nil {
		t.Error("expected client to be set")
	}
}

func TestNewFromEnv_MissingAPIKey(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")

	_, err := NewFromEnv()
	if err == nil {
		t.Error("expected error when both API keys are empty")
	}
}

func TestEmbed_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/embeddings" && r.Method == http.MethodPost {
			// Verify OpenRouter headers
			if r.Header.Get("HTTP-Referer") != "https://botla.app" {
				t.Error("missing or incorrect HTTP-Referer header")
			}
			if r.Header.Get("X-Title") != "Botla" {
				t.Error("missing or incorrect X-Title header")
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"data": [
					{
						"embedding": [0.7, 0.8, 0.9]
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

	embedder := NewEmbedder("test-key", server.URL, "text-embedding-3-small", server.Client())
	ctx := context.Background()

	result, err := embedder.Embed(ctx, "test text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 dimensions, got %d", len(result))
	}
	if result[0] != 0.7 || result[1] != 0.8 || result[2] != 0.9 {
		t.Errorf("unexpected embedding values: %v", result)
	}
}

func TestEmbedBatch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/embeddings" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"data": [
					{
						"embedding": [0.1, 0.2]
					},
					{
						"embedding": [0.3, 0.4]
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

	embedder := NewEmbedder("test-key", server.URL, "text-embedding-3-small", server.Client())
	ctx := context.Background()

	results, err := embedder.EmbedBatch(ctx, []string{"text 1", "text 2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if len(results[0]) != 2 || len(results[1]) != 2 {
		t.Error("unexpected embedding dimensions")
	}
	if results[0][0] != 0.1 || results[1][0] != 0.3 {
		t.Errorf("unexpected embedding values: %v", results)
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
			embedder := NewEmbedder("test-key", "https://openrouter.ai/api/v1", tt.model, nil)
			dim := embedder.Dimension()
			if dim != tt.expected {
				t.Errorf("expected dimension %d, got %d", tt.expected, dim)
			}
		})
	}
}
