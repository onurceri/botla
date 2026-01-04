package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/pkg/config"
)

func TestNewOpenRouterClient(t *testing.T) {
	t.Run("MissingKeys", func(t *testing.T) {
		cfg := &config.Config{}
		client, err := NewOpenRouterClient(cfg)
		if err == nil {
			t.Error("Expected error when no keys are set, got nil")
		}
		if client != nil {
			t.Error("Expected nil client when no keys are set")
		}
	})

	t.Run("OpenRouterKey", func(t *testing.T) {
		cfg := &config.Config{
			OPENROUTER_API_KEY: "sk-or-test",
		}
		client, err := NewOpenRouterClient(cfg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client == nil {
			t.Fatal("Expected client, got nil")
		}
		if client.apiKey != "sk-or-test" {
			t.Errorf("Expected apiKey 'sk-or-test', got '%s'", client.apiKey)
		}
	})

	t.Run("FallbackOpenAIKey", func(t *testing.T) {
		cfg := &config.Config{
			OPENAI_API_KEY: "sk-openai-test",
		}
		client, err := NewOpenRouterClient(cfg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client.apiKey != "sk-openai-test" {
			t.Errorf("Expected apiKey 'sk-openai-test', got '%s'", client.apiKey)
		}
	})

	t.Run("CustomBaseURL", func(t *testing.T) {
		cfg := &config.Config{
			OPENROUTER_API_KEY:  "test",
			OPENROUTER_API_BASE: "https://custom.api/v1",
		}
		client, err := NewOpenRouterClient(cfg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client.base != "https://custom.api/v1" {
			t.Errorf("Expected base 'https://custom.api/v1', got '%s'", client.base)
		}
	})

	t.Run("CustomTimeout", func(t *testing.T) {
		cfg := &config.Config{
			OPENROUTER_API_KEY:    "test",
			OPENROUTER_TIMEOUT_MS: 5000,
		}
		client, err := NewOpenRouterClient(cfg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client.http.Timeout != 5*time.Second {
			t.Errorf("Expected timeout 5s, got %v", client.http.Timeout)
		}
	})
}

func TestOpenRouterClient_CreateEmbedding(t *testing.T) {
	// Mock server
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Errorf("Expected path /embeddings, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("Expected Bearer test-key, got %s", auth)
		}

		var req embeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		resp := embeddingResponse{
			Data: []struct {
				Embedding []float64 `json:"embedding"`
			}{
				{Embedding: []float64{0.1, 0.2, 0.3}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := &OpenRouterClient{
		apiKey: "test-key",
		base:   server.URL,
		http:   server.Client(),
	}

	ctx := context.Background()
	emb, err := client.CreateEmbedding(ctx, "hello world")
	if err != nil {
		t.Fatalf("CreateEmbedding failed: %v", err)
	}
	if len(emb) != 3 {
		t.Errorf("Expected embedding len 3, got %d", len(emb))
	}
	if emb[0] != 0.1 {
		t.Errorf("Expected emb[0] 0.1, got %f", emb[0])
	}
}

func TestOpenRouterClient_CreateEmbedding_RetryAndFail(t *testing.T) {
	calls := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := &OpenRouterClient{
		apiKey: "test-key",
		base:   server.URL,
		http:   server.Client(),
	}

	ctx := context.Background()
	_, err := client.CreateEmbedding(ctx, "fail")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if calls != 4 {
		t.Errorf("Expected 4 calls, got %d", calls)
	}
}

func TestOpenRouterClient_CreateEmbeddingsBatch(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		resp := embeddingResponse{
			Data: []struct {
				Embedding []float64 `json:"embedding"`
			}{
				{Embedding: []float64{0.1, 0.2}},
				{Embedding: []float64{0.3, 0.4}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := &OpenRouterClient{
		apiKey: "test-key",
		base:   server.URL,
		http:   server.Client(),
	}

	ctx := context.Background()
	embs, err := client.CreateEmbeddingsBatch(ctx, []string{"one", "two"})
	if err != nil {
		t.Fatalf("CreateEmbeddingsBatch failed: %v", err)
	}
	if len(embs) != 2 {
		t.Errorf("Expected 2 embeddings, got %d", len(embs))
	}
	if embs[0][0] != 0.1 || embs[1][1] != 0.4 {
		t.Error("Embedding values incorrect")
	}

	// Test empty
	embsEmpty, err := client.CreateEmbeddingsBatch(ctx, nil)
	if err != nil {
		t.Error("Expected no error for empty input")
	}
	if embsEmpty != nil {
		t.Error("Expected nil result for empty input")
	}
}

func TestOpenRouterClient_CreateCompletion(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Verify fields
		if req.Model != "my-model" {
			t.Errorf("Expected model my-model, got %s", req.Model)
		}

		msg := "Hello human"
		resp := chatResponse{
			Choices: []struct {
				Message ChatMessage `json:"message"`
			}{
				{
					Message: ChatMessage{Content: &msg},
				},
			},
			Usage: struct {
				TotalTokens int `json:"total_tokens"`
			}{TotalTokens: 42},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := &OpenRouterClient{
		apiKey:       "test-key",
		base:         server.URL,
		defaultModel: "default-model",
		http:         server.Client(),
	}

	ctx := context.Background()
	params := models.CompletionParams{
		Model:        "my-model",
		SystemPrompt: "sys",
		UserMessage:  "hi",
		Context:      "ctx",
		MaxTokens:    100,
		Temperature:  0.7,
	}

	res, err := client.CreateCompletion(ctx, params)
	if err != nil {
		t.Fatalf("CreateCompletion failed: %v", err)
	}
	if res.Content != "Hello human" {
		t.Errorf("Expected 'Hello human', got '%s'", res.Content)
	}
	if res.UsageTokens != 42 {
		t.Errorf("Expected 42 tokens, got %d", res.UsageTokens)
	}
}

func TestOpenRouterClient_CreateCompletion_DefaultModel(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "default-model" {
			t.Errorf("Expected default-model, got %s", req.Model)
		}
		// Return minimal valid response
		msg := "ok"
		resp := chatResponse{
			Choices: []struct {
				Message ChatMessage `json:"message"`
			}{
				{
					Message: ChatMessage{Content: &msg},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := &OpenRouterClient{
		apiKey:       "test-key",
		base:         server.URL,
		defaultModel: "default-model",
		http:         server.Client(),
	}

	ctx := context.Background()
	// No Model specified in params
	params := models.CompletionParams{
		UserMessage: "hi",
	}
	_, err := client.CreateCompletion(ctx, params)
	if err != nil {
		t.Errorf("Failed: %v", err)
	}
}

func TestOpenRouterClient_GetModelInfo(t *testing.T) {
	client := &OpenRouterClient{}
	info := client.GetModelInfo()
	if info.Provider != "openrouter" {
		t.Errorf("Expected provider openrouter, got %s", info.Provider)
	}
	if info.MaxTokens == 0 {
		t.Error("Expected MaxTokens > 0")
	}
}

func TestOpenRouterClient_CreateCompletionWithTools_IncludesErrorBody(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid tools"}}`))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := &OpenRouterClient{
		apiKey: "test-key",
		base:   server.URL,
		http:   server.Client(),
	}

	ctx := context.Background()
	msg := "hi"
	_, err := client.CreateCompletionWithTools(ctx, []ChatMessage{{Role: "user", Content: &msg}}, []Tool{{Type: "function", Function: ToolFunction{Name: "list_sources", Description: "d", Parameters: json.RawMessage(`{"type":"object"}`)}}}, "openai/gpt-4o-mini", 0.1, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() == "OpenRouter error: 400 Bad Request" {
		t.Fatalf("expected error to include response body, got: %s", err.Error())
	}
}

func TestNewOpenRouterClient_WithHTTPClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			msg := "Hello"
			resp := chatResponse{
				Choices: []struct {
					Message ChatMessage `json:"message"`
				}{
					{Message: ChatMessage{Content: &msg}},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	mockHTTPClient := srv.Client()
	mockHTTPClient.Timeout = 10 * time.Second

	cfg := &config.Config{
		OPENROUTER_API_KEY:  "test-key",
		OPENROUTER_API_BASE: srv.URL,
	}

	client, err := NewOpenRouterClient(cfg, WithOpenRouterHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client, got nil")
	}
	if client.http != mockHTTPClient {
		t.Error("HTTP client was not injected correctly")
	}

	ctx := context.Background()
	params := models.CompletionParams{
		Model:       "test-model",
		UserMessage: "hi",
	}
	res, err := client.CreateCompletion(ctx, params)
	if err != nil {
		t.Fatalf("CreateCompletion failed: %v", err)
	}
	if res.Content != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", res.Content)
	}
}

func TestNewOpenRouterClient_WithNilHTTPClient(t *testing.T) {
	cfg := &config.Config{
		OPENROUTER_API_KEY: "test-key",
	}

	client, err := NewOpenRouterClient(cfg, WithOpenRouterHTTPClient(nil))
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

func TestNewOpenRouterClient_WithoutOptions(t *testing.T) {
	cfg := &config.Config{
		OPENROUTER_API_KEY: "test-key",
	}

	client, err := NewOpenRouterClient(cfg)
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

func TestNewOpenRouterClient_MultipleOptions(t *testing.T) {
	mockHTTPClient := &http.Client{Timeout: 5 * time.Second}
	cfg := &config.Config{
		OPENROUTER_API_KEY: "test-key",
	}

	client, err := NewOpenRouterClient(cfg, WithOpenRouterHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if client.http != mockHTTPClient {
		t.Error("HTTP client was not injected correctly with multiple options")
	}
}
