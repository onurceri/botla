package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/config"
)

type OpenRouterClient struct {
	apiKey       string
	http         *http.Client
	base         string
	defaultModel string
}

func NewOpenRouterClientFromEnv() (*OpenRouterClient, error) {
	k := os.Getenv("OPENROUTER_API_KEY")
	// Fallback to OPENAI_API_KEY if OPENROUTER_API_KEY is missing.
	if k == "" {
		k = os.Getenv("OPENAI_API_KEY")
	}
	if k == "" {
		return nil, errors.New("OPENROUTER_API_KEY (or OPENAI_API_KEY) is empty")
	}

	b := os.Getenv("OPENROUTER_API_BASE")
	if b == "" {
		b = "https://openrouter.ai/api/v1"
	}
	to := 30 * time.Second
	if v := os.Getenv("OPENROUTER_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			to = time.Duration(n) * time.Millisecond
		}
	}

	defModel := config.DefaultChatbotModel()

	return &OpenRouterClient{
		apiKey:       k,
		http:         &http.Client{Timeout: to},
		base:         b,
		defaultModel: defModel,
	}, nil
}

func (c *OpenRouterClient) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Use standard OpenAI-compatible embedding endpoint.
	body := embeddingRequest{Model: "text-embedding-3-small", Input: text}
	b, _ := json.Marshal(body)
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/embeddings", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		// OpenRouter specific headers
		req.Header.Set("HTTP-Referer", "https://botla.co") // Placeholder
		req.Header.Set("X-Title", "Botla")

		res, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			if res.StatusCode == http.StatusOK {
				var er embeddingResponse
				err := json.NewDecoder(res.Body).Decode(&er)
				_ = res.Body.Close()
				if err != nil {
					lastErr = err
				} else if len(er.Data) == 0 {
					lastErr = errors.New("no embedding returned")
				} else {
					out := make([]float32, len(er.Data[0].Embedding))
					for i, v := range er.Data[0].Embedding {
						out[i] = float32(v)
					}
					return out, nil
				}
			} else {
				lastErr = errors.New(res.Status)
				_ = res.Body.Close()
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}

func (c *OpenRouterClient) CreateEmbeddingsBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body := embeddingBatchRequest{Model: "text-embedding-3-small", Input: texts}
	b, _ := json.Marshal(body)
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/embeddings", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HTTP-Referer", "https://botla.co")
		req.Header.Set("X-Title", "Botla")

		res, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			if res.StatusCode == http.StatusOK {
				var er embeddingResponse
				err := json.NewDecoder(res.Body).Decode(&er)
				_ = res.Body.Close()
				if err != nil {
					lastErr = err
				} else if len(er.Data) == 0 {
					lastErr = errors.New("no embedding returned")
				} else {
					out := make([][]float32, len(er.Data))
					for i := range er.Data {
						out[i] = make([]float32, len(er.Data[i].Embedding))
						for j, v := range er.Data[i].Embedding {
							out[i][j] = float32(v)
						}
					}
					return out, nil
				}
			} else {
				lastErr = errors.New(res.Status)
				_ = res.Body.Close()
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}

func (c *OpenRouterClient) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{
		Name:              "openai/gpt-3.5-turbo", // Default guess
		Provider:          "openrouter",
		MaxTokens:         128000,
		SupportedFeatures: []string{"chat", "tools"},
	}
}

func (c *OpenRouterClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	model := params.Model
	if model == "" {
		model = c.defaultModel
	}
	// Assume caller provides correct model ID (e.g. "openai/gpt-4").
	user := "Context:\n" + params.Context + "\n\nQuestion:\n" + params.UserMessage
	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: params.SystemPrompt},
			{Role: "user", Content: user},
		},
		Temperature: params.Temperature,
		MaxTokens:   params.MaxTokens,
	}
	b, _ := json.Marshal(reqBody)
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/chat/completions", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HTTP-Referer", "https://botla.co")
		req.Header.Set("X-Title", "Botla")

		res, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			if res.StatusCode == http.StatusOK {
				var cr chatResponse
				if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
					lastErr = err
				} else {
					_ = res.Body.Close()
					if len(cr.Choices) > 0 {
						return &models.CompletionResult{
							Content:     cr.Choices[0].Message.Content,
							UsageTokens: cr.Usage.TotalTokens,
						}, nil
					}
					lastErr = errors.New("no choices in response")
				}
			} else {
				lastErr = errors.New("OpenRouter error: " + res.Status)
				_ = res.Body.Close()
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}
