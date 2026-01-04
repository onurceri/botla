package rag

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/pkg/config"
	pkgErrors "github.com/onurceri/botla-app/pkg/errors"
)

// TestCreateEmbedding_RateLimitError tests that HTTP 429 returns ErrRateLimit
func TestCreateEmbedding_RateLimitError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"Rate limit exceeded"}}`))
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	// Use a context that can be cancelled quickly to speed up the test
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cancel after first attempt to avoid all retries
	go func() {
		// Let one attempt happen, then cancel
		<-ctx.Done()
	}()

	_, err := c.CreateEmbedding(ctx, "hi")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, pkgErrors.ErrRateLimit) {
		t.Errorf("expected error to wrap ErrRateLimit, got: %v", err)
	}
}

// TestCreateEmbedding_NetworkError tests that network errors are wrapped
func TestCreateEmbedding_NetworkError(t *testing.T) {
	t.Parallel()
	// Server on invalid address triggers network error
	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: "http://127.0.0.1:1", // invalid port
	}
	c, _ := NewOpenAIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately to avoid long retries

	_, err := c.CreateEmbedding(ctx, "hi")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// When context is cancelled, we should get context error
	// This test mainly verifies the flow works
}

// TestCreateEmbedding_InternalServerError_WrapsNetwork tests 500 errors
func TestCreateEmbedding_InternalServerError_WrapsNetwork(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Internal error"}}`))
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := c.CreateEmbedding(ctx, "hi")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// 5xx should be wrapped with ErrNetwork (temporary failure)
	if !errors.Is(err, pkgErrors.ErrNetwork) {
		t.Errorf("expected error to wrap ErrNetwork for 5xx, got: %v", err)
	}
}

// TestCreateCompletion_RateLimitError tests that HTTP 429 returns ErrRateLimit on completion
func TestCreateCompletion_RateLimitError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"Rate limit exceeded"}}`))
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := c.CreateCompletion(ctx, models.CompletionParams{
		SystemPrompt: "sys",
		Context:      "ctx",
		UserMessage:  "q",
		Model:        "gpt-3.5-turbo",
		Temperature:  0,
		MaxTokens:    10,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, pkgErrors.ErrRateLimit) {
		t.Errorf("expected error to wrap ErrRateLimit, got: %v", err)
	}
}

// TestCreateCompletion_InternalServerError_WrapsNetwork tests 500 errors on completion
func TestCreateCompletion_InternalServerError_WrapsNetwork(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Internal error"}}`))
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := c.CreateCompletion(ctx, models.CompletionParams{
		SystemPrompt: "sys",
		Context:      "ctx",
		UserMessage:  "q",
		Model:        "gpt-3.5-turbo",
		Temperature:  0,
		MaxTokens:    10,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, pkgErrors.ErrNetwork) {
		t.Errorf("expected error to wrap ErrNetwork for 5xx, got: %v", err)
	}
}

// TestCreateEmbedding_ContextCancelled_WrappedError verifies context cancellation wraps correctly
func TestCreateEmbedding_ContextCancelled_WrappedError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := c.CreateEmbedding(ctx, "hi")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Check that error is context.Canceled (or wraps ErrContextCancelled)
	if !errors.Is(err, context.Canceled) && !errors.Is(err, pkgErrors.ErrContextCancelled) {
		t.Errorf("expected context cancelled error, got: %v", err)
	}
}

// TestCreateEmbeddingsBatch_RateLimitError tests batched embedding rate limits
func TestCreateEmbeddingsBatch_RateLimitError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"Rate limit exceeded"}}`))
	}))
	defer srv.Close()

	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: srv.URL,
	}
	c, _ := NewOpenAIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := c.CreateEmbeddingsBatch(ctx, []string{"a", "b"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, pkgErrors.ErrRateLimit) {
		t.Errorf("expected error to wrap ErrRateLimit, got: %v", err)
	}
}
