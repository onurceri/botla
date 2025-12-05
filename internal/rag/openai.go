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

	"github.com/onurceri/botla-co/pkg/config"
)

type OpenAIClient struct {
	apiKey       string
	http         *http.Client
	base         string
	defaultModel string
}

func NewOpenAIClientFromEnv() (*OpenAIClient, error) {
	k := os.Getenv("OPENAI_API_KEY")
	if k == "" {
		return nil, errors.New("OPENAI_API_KEY is empty")
	}
	b := os.Getenv("OPENAI_API_BASE")
	if b == "" {
		b = "https://api.openai.com"
	}
	to := 30 * time.Second
	if v := os.Getenv("OPENAI_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			to = time.Duration(n) * time.Millisecond
		}
	}

	defModel := config.DefaultChatbotModel()
	return &OpenAIClient{
		apiKey:       k,
		http:         &http.Client{Timeout: to},
		base:         b,
		defaultModel: defModel,
	}, nil
}

// Embeddings
type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
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

func (c *OpenAIClient) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	body := embeddingRequest{Model: "text-embedding-3-small", Input: text}
	b, _ := json.Marshal(body)
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/embeddings", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			defer func() { _ = res.Body.Close() }()
			if res.StatusCode == http.StatusOK {
				var er embeddingResponse
				if err := json.NewDecoder(res.Body).Decode(&er); err != nil {
					lastErr = err
				} else {
					if len(er.Data) == 0 {
						lastErr = errors.New("no embedding returned")
					} else {
						out := make([]float32, len(er.Data[0].Embedding))
						for i, v := range er.Data[0].Embedding {
							out[i] = float32(v)
						}
						return out, nil
					}
				}
			} else {
				lastErr = errors.New(res.Status)
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}

// Batch Embeddings
type embeddingBatchRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

func (c *OpenAIClient) CreateEmbeddingsBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body := embeddingBatchRequest{Model: "text-embedding-3-small", Input: texts}
	b, _ := json.Marshal(body)
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/embeddings", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			defer func() { _ = res.Body.Close() }()
			if res.StatusCode == http.StatusOK {
				var er embeddingResponse
				if err := json.NewDecoder(res.Body).Decode(&er); err != nil {
					lastErr = err
				} else {
					if len(er.Data) == 0 {
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
				}
			} else {
				lastErr = errors.New(res.Status)
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}

// Completions
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float32       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *OpenAIClient) CreateCompletion(ctx context.Context, systemPrompt, contextText, userMessage string, model string, temperature float32, maxTokens int) (string, int, error) {
	if model == "" {
		model = c.defaultModel
	}
	user := "Context:\n" + contextText + "\n\nQuestion:\n" + userMessage
	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: user},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	b, _ := json.Marshal(reqBody)
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/chat/completions", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			defer func() { _ = res.Body.Close() }()
			if res.StatusCode == http.StatusOK {
				var cr chatResponse
				if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
					lastErr = err
				} else {
					if len(cr.Choices) == 0 {
						lastErr = errors.New("no completion returned")
					} else {
						return cr.Choices[0].Message.Content, cr.Usage.TotalTokens, nil
					}
				}
			} else {
				lastErr = errors.New(res.Status)
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return "", 0, lastErr
}
