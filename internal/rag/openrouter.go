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

// NewOpenRouterClient creates an OpenRouter client from config
func NewOpenRouterClient(cfg *config.Config) (*OpenRouterClient, error) {
	k := ""
	if cfg != nil {
		k = cfg.OPENROUTER_API_KEY
		// Fallback to OPENAI_API_KEY if OPENROUTER_API_KEY is missing.
		if k == "" {
			k = cfg.OPENAI_API_KEY
		}
	}
	if k == "" {
		return nil, errors.New("OPENROUTER_API_KEY (or OPENAI_API_KEY) is empty")
	}

	base := ""
	if cfg != nil && cfg.OPENROUTER_API_BASE != "" {
		base = cfg.OPENROUTER_API_BASE
	} else {
		base = os.Getenv("OPENROUTER_API_BASE")
	}
	if base == "" {
		base = "https://openrouter.ai/api/v1"
	}

	to := 30 * time.Second
	if cfg != nil && cfg.OPENROUTER_TIMEOUT_MS > 0 {
		to = time.Duration(cfg.OPENROUTER_TIMEOUT_MS) * time.Millisecond
	}

	defModel := config.DefaultChatbotModel()
	if cfg != nil && cfg.DEFAULT_CHATBOT_MODEL != "" {
		defModel = cfg.DEFAULT_CHATBOT_MODEL
	}

	httpClient := &http.Client{Timeout: to}
	if GlobalHTTPClient != nil {
		httpClient = GlobalHTTPClient
	}

	return &OpenRouterClient{
		apiKey:       k,
		http:         httpClient,
		base:         base,
		defaultModel: defModel,
	}, nil
}

func (c *OpenRouterClient) getHTTPClient() *http.Client {
	if GlobalHTTPClient != nil {
		return GlobalHTTPClient
	}
	return c.http
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
		req.Header.Set("HTTP-Referer", "https://botla.app") // Placeholder
		req.Header.Set("X-Title", "Botla")

		res, err := c.getHTTPClient().Do(req)
		switch {
		case err != nil:
			lastErr = fmt.Errorf("http post embedding: %w", err)
		case res.StatusCode != http.StatusOK:
			lastErr = fmt.Errorf("http post embedding status: %s", res.Status)
			_ = res.Body.Close()
		default:
			var er embeddingResponse
			err := json.NewDecoder(res.Body).Decode(&er)
			_ = res.Body.Close()
			switch {
			case err != nil:
				lastErr = fmt.Errorf("decode embedding response: %w", err)
			case len(er.Data) == 0:
				lastErr = fmt.Errorf("no embedding returned")
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
		req.Header.Set("HTTP-Referer", "https://botla.app")
		req.Header.Set("X-Title", "Botla")

		res, err := c.getHTTPClient().Do(req)
		switch {
		case err != nil:
			lastErr = fmt.Errorf("http post batch embedding: %w", err)
		case res.StatusCode != http.StatusOK:
			lastErr = fmt.Errorf("http post batch embedding status: %s", res.Status)
			_ = res.Body.Close()
		default:
			var er embeddingResponse
			err := json.NewDecoder(res.Body).Decode(&er)
			_ = res.Body.Close()
			switch {
			case err != nil:
				lastErr = fmt.Errorf("decode batch embedding response: %w", err)
			case len(er.Data) == 0:
				lastErr = fmt.Errorf("no batch embedding returned")
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
	reqBody := ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: &params.SystemPrompt},
			{Role: "user", Content: &user},
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
		req.Header.Set("HTTP-Referer", "https://botla.app")
		req.Header.Set("X-Title", "Botla")

		res, err := c.getHTTPClient().Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http post chat completion: %w", err)
		} else {
			if res.StatusCode == http.StatusOK {
				var cr chatResponse
				if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
					lastErr = fmt.Errorf("decode chat response: %w", err)
				} else {
					_ = res.Body.Close()
					if len(cr.Choices) > 0 {
						content := ""
						if cr.Choices[0].Message.Content != nil {
							content = *cr.Choices[0].Message.Content
						}
						return &models.CompletionResult{
							Content:     content,
							UsageTokens: cr.Usage.TotalTokens,
						}, nil
					}
					lastErr = fmt.Errorf("no choices in response")
				}
			} else {
				lastErr = fmt.Errorf("OpenRouter error: %s", res.Status)
				_ = res.Body.Close()
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}

// CreateCompletionWithTools sends a completion request with tool support
// OpenRouter uses OpenAI-compatible API format for tools
func (c *OpenRouterClient) CreateCompletionWithTools(
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
		return nil, fmt.Errorf("marshal tool request: %w", err)
	}
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/chat/completions", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HTTP-Referer", "https://botla.app")
		req.Header.Set("X-Title", "Botla")

		res, err := c.getHTTPClient().Do(req)
		switch {
		case err != nil:
			lastErr = fmt.Errorf("http post chat with tools: %w", err)
		case res.StatusCode != http.StatusOK:
			body, _ := io.ReadAll(io.LimitReader(res.Body, 8192))
			_ = res.Body.Close()
			if len(body) > 0 {
				lastErr = fmt.Errorf("openrouter error: %s: %s", res.Status, string(body))
			} else {
				lastErr = fmt.Errorf("openrouter error status: %s", res.Status)
			}
		default:
			var cr ChatResponseWithTools
			err := json.NewDecoder(res.Body).Decode(&cr)
			_ = res.Body.Close()
			switch {
			case err != nil:
				lastErr = fmt.Errorf("decode chat with tools response: %w", err)
			case len(cr.Choices) == 0:
				lastErr = fmt.Errorf("no completion returned")
			default:
				return &cr, nil
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}
