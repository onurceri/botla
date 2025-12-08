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

func (c *OpenAIClient) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{
		Name:              "gpt-4o-mini",
		Provider:          "openai",
		MaxTokens:         128000,
		SupportedFeatures: []string{"chat", "tools"},
	}
}

func (c *OpenAIClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	model := params.Model
	if model == "" {
		model = c.defaultModel
	}
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
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/chat/completions", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			if res.StatusCode == http.StatusOK {
				var cr chatResponse
				err := json.NewDecoder(res.Body).Decode(&cr)
				_ = res.Body.Close()
				if err != nil {
					lastErr = err
				} else if len(cr.Choices) == 0 {
					lastErr = errors.New("no completion returned")
				} else {
					return &models.CompletionResult{
						Content:     cr.Choices[0].Message.Content,
						UsageTokens: cr.Usage.TotalTokens,
					}, nil
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

type ChatRequestWithTools struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Tools       []Tool        `json:"tools,omitempty"`
	ToolChoice  string        `json:"tool_choice,omitempty"` // "auto", "none"
	Temperature float32       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatResponseWithTools struct {
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// CreateCompletionWithTools sends a completion request with tool support
func (c *OpenAIClient) CreateCompletionWithTools(
	ctx context.Context,
	messages []ChatMessage,
	tools []Tool,
	model string,
	temperature float32,
	maxTokens int,
) (*ChatResponseWithTools, error) {
	if model == "" {
		model = c.defaultModel
	}

	reqBody := ChatRequestWithTools{
		Model:       model,
		Messages:    messages,
		Tools:       tools,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	if len(tools) > 0 {
		reqBody.ToolChoice = "auto"
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
			if res.StatusCode == http.StatusOK {
				var cr ChatResponseWithTools
				err := json.NewDecoder(res.Body).Decode(&cr)
				_ = res.Body.Close()
				if err != nil {
					lastErr = err
				} else if len(cr.Choices) == 0 {
					lastErr = errors.New("no completion returned")
				} else {
					return &cr, nil
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
