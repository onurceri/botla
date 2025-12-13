package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// NewOpenAIClient creates an OpenAI client from config
func NewOpenAIClient(cfg *config.Config) (*OpenAIClient, error) {
	if cfg == nil || cfg.OPENAI_API_KEY == "" {
		return nil, errors.New("OPENAI_API_KEY is empty")
	}
	base := cfg.OPENAI_API_BASE
	if base == "" {
		base = "https://api.openai.com"
	}
	to := time.Duration(cfg.OPENAI_TIMEOUT_MS) * time.Millisecond
	if to <= 0 {
		to = 30 * time.Second
	}
	defModel := cfg.DEFAULT_CHATBOT_MODEL
	if defModel == "" {
		defModel = config.ModelGPT4oMini
	}
	return &OpenAIClient{
		apiKey:       cfg.OPENAI_API_KEY,
		http:         &http.Client{Timeout: to},
		base:         base,
		defaultModel: defModel,
	}, nil
}

// NewOpenAIClientFromEnv creates an OpenAI client from environment variables (backward compatibility)
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
		Name:              config.ModelGPT4oMini,
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
		switch {
		case err != nil:
			lastErr = err
		case res.StatusCode != http.StatusOK:
			lastErr = errors.New(res.Status)
			_ = res.Body.Close()
		default:
			var cr chatResponse
			err := json.NewDecoder(res.Body).Decode(&cr)
			_ = res.Body.Close()
			switch {
			case err != nil:
				lastErr = err
			case len(cr.Choices) == 0:
				lastErr = errors.New("no completion returned")
			default:
				return &models.CompletionResult{
					Content:     cr.Choices[0].Message.Content,
					UsageTokens: cr.Usage.TotalTokens,
				}, nil
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

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/chat/completions", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		switch {
		case err != nil:
			lastErr = err
		case res.StatusCode != http.StatusOK:
			body, _ := io.ReadAll(io.LimitReader(res.Body, 8192))
			_ = res.Body.Close()
			if len(body) > 0 {
				lastErr = fmt.Errorf("openai error: %s: %s", res.Status, string(body))
			} else {
				lastErr = errors.New(res.Status)
			}
		default:
			var cr ChatResponseWithTools
			err := json.NewDecoder(res.Body).Decode(&cr)
			_ = res.Body.Close()
			switch {
			case err != nil:
				lastErr = err
			case len(cr.Choices) == 0:
				lastErr = errors.New("no completion returned")
			default:
				return &cr, nil
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}
