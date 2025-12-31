package openai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/onurceri/botla-co/internal/ai"
)

// Config holds configuration for OpenAI embedder
type Config struct {
	APIKey  string        // Required
	BaseURL string        // Optional, defaults to https://api.openai.com
	Model   string        // Optional, defaults to text-embedding-3-small
	Timeout time.Duration // Optional, defaults to 30s
}

// Embedder implements ai.Embedder for OpenAI
type Embedder struct {
	client *ai.BaseClient
	model  string
}

// Verify interface compliance at compile time
var _ ai.Embedder = (*Embedder)(nil)

// NewEmbedder creates a new OpenAI embedder
func NewEmbedder(config Config, client *http.Client) (*Embedder, error) {
	if config.APIKey == "" {
		return nil, errors.New("api key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com"
	}
	if config.Model == "" {
		config.Model = "text-embedding-3-small"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if client == nil {
		client = &http.Client{Timeout: config.Timeout}
	}
	return &Embedder{
		client: ai.NewBaseClientWithHTTPClient(config.BaseURL, config.APIKey, nil, client),
		model:  config.Model,
	}, nil
}

// Embed generates an embedding for a single text
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if e == nil {
		return nil, fmt.Errorf("openai embedder is nil")
	}

	body := embeddingRequest{Model: e.model, Input: text}
	var resp embeddingResponse

	if err := e.client.Post(ctx, "/v1/embeddings", body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no embedding data returned")
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
		return nil, fmt.Errorf("openai embedder is nil")
	}
	if len(texts) == 0 {
		return nil, nil
	}

	body := embeddingBatchRequest{Model: e.model, Input: texts}
	var resp embeddingResponse

	if err := e.client.Post(ctx, "/v1/embeddings", body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no embedding data returned")
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

// Internal types for OpenAI API

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
