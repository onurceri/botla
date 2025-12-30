package openrouter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/onurceri/botla-co/internal/ai"
)

func init() {
	// Register OpenRouter embedder factory
	ai.RegisterEmbedder(ai.ProviderOpenRouter, func() (ai.Embedder, error) {
		return NewFromEnv()
	})
}

// Embedder implements ai.Embedder for OpenRouter
type Embedder struct {
	client *ai.BaseClient
	model  string
}

// Verify interface compliance at compile time
var _ ai.Embedder = (*Embedder)(nil)

// NewEmbedder creates a new OpenRouter embedder
func NewEmbedder(apiKey, baseURL string, model string, client *http.Client) *Embedder {
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	if model == "" {
		model = "text-embedding-3-small"
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	headers := map[string]string{
		"HTTP-Referer": "https://botla.app",
		"X-Title":      "Botla",
	}

	return &Embedder{
		client: ai.NewBaseClientWithHTTPClient(baseURL, apiKey, headers, client),
		model:  model,
	}
}

// NewFromEnv creates an OpenRouter embedder from environment variables
func NewFromEnv() (*Embedder, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("OPENROUTER_API_KEY (or OPENAI_API_KEY) is empty")
	}

	baseURL := os.Getenv("OPENROUTER_API_BASE")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	model := os.Getenv("OPENROUTER_EMBEDDING_MODEL")
	if model == "" {
		model = "text-embedding-3-small"
	}

	timeout := 30 * time.Second
	if timeoutStr := os.Getenv("OPENROUTER_TIMEOUT_MS"); timeoutStr != "" {
		if ms, err := strconv.Atoi(timeoutStr); err == nil && ms > 0 {
			timeout = time.Duration(ms) * time.Millisecond
		}
	}

	headers := map[string]string{
		"HTTP-Referer": "https://botla.app",
		"X-Title":      "Botla",
	}

	return &Embedder{
		client: ai.NewBaseClientWithHTTPClient(baseURL, apiKey, headers, &http.Client{Timeout: timeout}),
		model:  model,
	}, nil
}

// Embed generates an embedding for a single text
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if e == nil {
		return nil, fmt.Errorf("openrouter embedder is nil")
	}

	body := embeddingRequest{Model: e.model, Input: text}
	var resp embeddingResponse

	if err := e.client.Post(ctx, "/embeddings", body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no embedding returned")
	}

	out := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		out[i] = float32(v)
	}
	return out, nil
}

// EmbedBatch generates embeddings for multiple texts
func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if e == nil {
		return nil, fmt.Errorf("openrouter embedder is nil")
	}
	if len(texts) == 0 {
		return nil, nil
	}

	body := embeddingBatchRequest{Model: e.model, Input: texts}
	var resp embeddingResponse

	if err := e.client.Post(ctx, "/embeddings", body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no embedding returned")
	}

	out := make([][]float32, len(resp.Data))
	for i := range resp.Data {
		out[i] = make([]float32, len(resp.Data[i].Embedding))
		for j, v := range resp.Data[i].Embedding {
			out[i][j] = float32(v)
		}
	}
	return out, nil
}

// Dimension returns the dimensionality of embeddings
func (e *Embedder) Dimension() int {
	// text-embedding-3-small has 1536 dimensions
	// text-embedding-3-large has 3072 dimensions
	if e.model == "text-embedding-3-large" {
		return 3072
	}
	return 1536
}

// Internal types for OpenRouter API

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingBatchRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}
