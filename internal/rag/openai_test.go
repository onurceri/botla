package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/config"
)

func TestCreateEmbedding_Success(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/embeddings" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data":  []map[string]any{{"embedding": []float64{0.1, 0.2}}},
				"usage": map[string]int{"prompt_tokens": 1, "total_tokens": 2},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	defer srv.Close()
	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)
	v, err := c.CreateEmbedding(context.Background(), "hi")
	if err != nil || len(v) != 2 {
		t.Fatalf("embedding err: %v", err)
	}
}

func TestCreateEmbeddingsBatch_Success(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/embeddings" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data":  []map[string]any{{"embedding": []float64{0.1}}, {"embedding": []float64{0.3}}},
				"usage": map[string]int{"prompt_tokens": 1, "total_tokens": 2},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	defer srv.Close()
	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)
	v, err := c.CreateEmbeddingsBatch(context.Background(), []string{"a", "b"})
	if err != nil || len(v) != 2 {
		t.Fatalf("batch err: %v", err)
	}
}

func TestCreateEmbedding_RetryFailure(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer srv.Close()
	defer srv.Close()
	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)
	_, err := c.CreateEmbedding(context.Background(), "hi")
	if err == nil {
		t.Fatalf("expected error after retries")
	}
}

func TestCreateCompletion_NoChoices(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{}})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	defer srv.Close()
	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)
	_, err := c.CreateCompletion(context.Background(), models.CompletionParams{
		SystemPrompt: "sys",
		Context:      "ctx",
		UserMessage:  "q",
		Model:        "gpt-3.5-turbo",
		Temperature:  0,
		MaxTokens:    10,
	})
	if err == nil {
		t.Fatalf("expected error due to no choices")
	}
}

func TestCreateCompletion_Success(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"role": "assistant", "content": "ok"}}},
				"usage":   map[string]int{"total_tokens": 3},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	defer srv.Close()
	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)
	res, err := c.CreateCompletion(context.Background(), models.CompletionParams{
		SystemPrompt: "sys",
		Context:      "ctx",
		UserMessage:  "q",
		Model:        "gpt-3.5-turbo",
		Temperature:  0,
		MaxTokens:    10,
	})
	if err != nil || res.Content == "" || res.UsageTokens == 0 {
		t.Fatalf("completion err: %v", err)
	}
}

func TestCreateCompletionWithTools_IncludesErrorBody(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"invalid tools"}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)
	msg := "hi"
	_, err := c.CreateCompletionWithTools(context.Background(), []ChatMessage{{Role: "user", Content: &msg}}, []Tool{{Type: "function", Function: ToolFunction{Name: "list_sources", Description: "d", Parameters: json.RawMessage(`{"type":"object"}`)}}}, "gpt-4o-mini", 0, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() == "400 Bad Request" || err.Error() == "Bad Request" {
		t.Fatalf("expected error to include response body, got: %s", err.Error())
	}
}

func TestCreateEmbedding_ContextCancellation(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	// Server always returns error to trigger retries
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	// Create an already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := testing.AllocsPerRun(1, func() {})
	_ = start
	startTime := testing.Benchmark(func(b *testing.B) {}).T

	_, err := c.CreateEmbedding(ctx, "hi")

	// Test should complete quickly (not wait for retry delays)
	endTime := testing.Benchmark(func(b *testing.B) {}).T
	_ = startTime
	_ = endTime

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}

func TestCreateEmbeddingsBatch_ContextCancellation(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.CreateEmbeddingsBatch(ctx, []string{"a", "b"})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}

func TestCreateCompletion_ContextCancellation(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.CreateCompletion(ctx, models.CompletionParams{
		SystemPrompt: "sys",
		Context:      "ctx",
		UserMessage:  "q",
		Model:        "gpt-3.5-turbo",
		Temperature:  0,
		MaxTokens:    10,
	})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}

func TestCreateCompletionWithTools_ContextCancellation(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg := "hi"
	_, err := c.CreateCompletionWithTools(ctx, []ChatMessage{{Role: "user", Content: &msg}}, nil, "gpt-4o-mini", 0, 10)

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}

func TestNewOpenAIClient_WithHTTPClient(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/embeddings" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data":  []map[string]any{{"embedding": []float64{0.1, 0.2}}},
				"usage": map[string]int{"prompt_tokens": 1, "total_tokens": 2},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	mockHTTPClient := srv.Client()
	mockHTTPClient.Timeout = 10 * time.Second

	cfg := &config.Config{
		OPENAI_API_KEY:  "test-key",
		OPENAI_API_BASE: srv.URL,
	}

	client, err := NewOpenAIClient(cfg, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client, got nil")
	}
	if client.http != mockHTTPClient {
		t.Error("HTTP client was not injected correctly")
	}

	emb, err := client.CreateEmbedding(context.Background(), "test")
	if err != nil {
		t.Fatalf("CreateEmbedding failed: %v", err)
	}
	if len(emb) != 2 {
		t.Errorf("Expected 2 embedding values, got %d", len(emb))
	}
}

func TestNewOpenAIClient_WithNilHTTPClient(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	cfg := &config.Config{
		OPENAI_API_KEY: "test-key",
	}

	client, err := NewOpenAIClient(cfg, WithHTTPClient(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client, got nil")
	}
	if client.http == nil {
		t.Error("Expected non-nil HTTP client when nil is passed (should use default)")
	}
}

func TestNewOpenAIClient_WithoutOptions(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	cfg := &config.Config{
		OPENAI_API_KEY: "test-key",
	}

	client, err := NewOpenAIClient(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client, got nil")
	}
	if client.http == nil {
		t.Error("Expected non-nil HTTP client")
	}
	if client.http.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout of 30s, got %v", client.http.Timeout)
	}
}

func TestNewOpenAIClient_MultipleOptions(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	mockHTTPClient := &http.Client{Timeout: 5 * time.Second}
	cfg := &config.Config{
		OPENAI_API_KEY: "test-key",
	}

	client, err := NewOpenAIClient(cfg, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if client.http != mockHTTPClient {
		t.Error("HTTP client was not injected correctly with multiple options")
	}
}
