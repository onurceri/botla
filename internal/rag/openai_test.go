package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
)

func TestNewOpenAIClientFromEnv(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	c, err := NewOpenAIClientFromEnv()
	if err != nil || c == nil {
		t.Fatalf("client err: %v", err)
	}
}

func TestNewOpenAIClientFromEnv_MissingKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	c, err := NewOpenAIClientFromEnv()
	if err == nil || c != nil {
		t.Fatalf("expected error for missing key")
	}
}

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
	t.Setenv("OPENAI_API_BASE", srv.URL)
	c, _ := NewOpenAIClientFromEnv()
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
	t.Setenv("OPENAI_API_BASE", srv.URL)
	c, _ := NewOpenAIClientFromEnv()
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
	t.Setenv("OPENAI_API_BASE", srv.URL)
	c, _ := NewOpenAIClientFromEnv()
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
	t.Setenv("OPENAI_API_BASE", srv.URL)
	c, _ := NewOpenAIClientFromEnv()
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
	t.Setenv("OPENAI_API_BASE", srv.URL)
	c, _ := NewOpenAIClientFromEnv()
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
	t.Setenv("OPENAI_API_BASE", srv.URL)

	c, _ := NewOpenAIClientFromEnv()
	msg := "hi"
	_, err := c.CreateCompletionWithTools(context.Background(), []ChatMessage{{Role: "user", Content: &msg}}, []Tool{{Type: "function", Function: ToolFunction{Name: "list_sources", Description: "d", Parameters: json.RawMessage(`{"type":"object"}`)}}}, "gpt-4o-mini", 0, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() == "400 Bad Request" || err.Error() == "Bad Request" {
		t.Fatalf("expected error to include response body, got: %s", err.Error())
	}
}
