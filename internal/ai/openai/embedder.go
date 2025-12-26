package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/onurceri/botla-co/internal/ai"
)

func init() {
	// Register OpenAI embedder factory
	ai.RegisterEmbedder(ai.ProviderOpenAI, func() (ai.Embedder, error) {
		return NewFromEnv()
	})
}


// Embedder implements ai.Embedder for OpenAI
type Embedder struct {
	apiKey string
	http   *http.Client
	base   string
	model  string
}

// Verify interface compliance at compile time
var _ ai.Embedder = (*Embedder)(nil)

// globalHTTPClient can be set in tests to override all clients
var globalHTTPClient *http.Client

// NewEmbedder creates a new OpenAI embedder
func NewEmbedder(apiKey, baseURL string, model string, client *http.Client) *Embedder {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	if model == "" {
		model = "text-embedding-3-small"
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &Embedder{
		apiKey: apiKey,
		http:   client,
		base:   baseURL,
		model:  model,
	}
}

// NewFromEnv creates an OpenAI embedder from environment variables
func NewFromEnv() (*Embedder, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is empty")
	}

	baseURL := os.Getenv("OPENAI_API_BASE")
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	model := os.Getenv("OPENAI_EMBEDDING_MODEL")
	if model == "" {
		model = "text-embedding-3-small"
	}

	timeout := 30 * time.Second
	// Check if OPENAI_TIMEOUT_MS is set
	if timeoutStr := os.Getenv("OPENAI_TIMEOUT_MS"); timeoutStr != "" {
		if ms, err := strconv.Atoi(timeoutStr); err == nil && ms > 0 {
			timeout = time.Duration(ms) * time.Millisecond
		}
	}

	return &Embedder{
		apiKey: apiKey,
		http:   &http.Client{Timeout: timeout},
		base:   baseURL,
		model:  model,
	}, nil
}

func (e *Embedder) getHTTPClient() *http.Client {
	if globalHTTPClient != nil {
		return globalHTTPClient
	}
	return e.http
}

// Embed generates an embedding for a single text
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if e == nil {
		return nil, fmt.Errorf("openai embedder is nil")
	}

	body := embeddingRequest{Model: e.model, Input: text}
	b, _ := json.Marshal(body)

	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, e.base+"/v1/embeddings", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
		req.Header.Set("Content-Type", "application/json")

		res, err := e.getHTTPClient().Do(req)
		switch {
		case err != nil:
			lastErr = err
		case res.StatusCode != http.StatusOK:
			lastErr = errors.New(res.Status)
			_ = res.Body.Close()
		default:
			var er embeddingResponse
			err := json.NewDecoder(res.Body).Decode(&er)
			_ = res.Body.Close()
			switch {
			case err != nil:
				lastErr = err
			case len(er.Data) == 0:
				lastErr = errors.New("no embedding returned")
			default:
				out := make([]float32, len(er.Data[0].Embedding))
				for i, v := range er.Data[0].Embedding {
					out[i] = float32(v)
				}
				return out, nil
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
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
	b, _ := json.Marshal(body)

	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, e.base+"/v1/embeddings", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
		req.Header.Set("Content-Type", "application/json")

		res, err := e.getHTTPClient().Do(req)
		switch {
		case err != nil:
			lastErr = err
		case res.StatusCode != http.StatusOK:
			lastErr = errors.New(res.Status)
			_ = res.Body.Close()
		default:
			var er embeddingResponse
			err := json.NewDecoder(res.Body).Decode(&er)
			_ = res.Body.Close()
			switch {
			case err != nil:
				lastErr = err
			case len(er.Data) == 0:
				lastErr = errors.New("no embedding returned")
			default:
				out := make([][]float32, len(er.Data))
				for i := range er.Data {
					out[i] = make([]float32, len(er.Data[i].Embedding))
					for j, v := range er.Data[i].Embedding {
						out[i][j] = float32(v)
					}
				}
				return out, nil
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
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
