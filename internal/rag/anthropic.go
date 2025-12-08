package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

type AnthropicClient struct {
	apiKey       string
	http         *http.Client
	base         string
	defaultModel string
}

func NewAnthropicClientFromEnv() (*AnthropicClient, error) {
	k := os.Getenv("ANTHROPIC_API_KEY")
	if k == "" {
		return nil, errors.New("ANTHROPIC_API_KEY is empty")
	}
	b := os.Getenv("ANTHROPIC_API_BASE")
	if b == "" {
		b = "https://api.anthropic.com"
	}

	return &AnthropicClient{
		apiKey:       k,
		http:         &http.Client{Timeout: 60 * time.Second},
		base:         b,
		defaultModel: "claude-3-5-sonnet-20241022",
	}, nil
}

func (c *AnthropicClient) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{
		Name:              "claude-3-5-sonnet-20241022",
		Provider:          "anthropic",
		MaxTokens:         200000,
		SupportedFeatures: []string{"chat"},
	}
}

// Anthropic Messages API structs
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float32            `json:"temperature,omitempty"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *AnthropicClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	model := params.Model
	if model == "" {
		model = c.defaultModel
	}
	// Strip "anthropic:" prefix if present (will be handled by factory later, but good for safety)
	if len(model) > 10 && model[:10] == "anthropic:" {
		model = model[10:]
	}

	userContent := "Context:\n" + params.Context + "\n\nQuestion:\n" + params.UserMessage

	reqBody := anthropicRequest{
		Model: model,
		Messages: []anthropicMessage{
			{Role: "user", Content: userContent},
		},
		System:      params.SystemPrompt,
		MaxTokens:   params.MaxTokens,
		Temperature: params.Temperature,
	}

	if reqBody.MaxTokens == 0 {
		reqBody.MaxTokens = 1024 // Default for Anthropic
	}

	b, _ := json.Marshal(reqBody)
	var lastErr error

	for attempt := 0; attempt < 4; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/messages", bytes.NewReader(b))
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("content-type", "application/json")

		res, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			if res.StatusCode == http.StatusOK {
				var ar anthropicResponse
				err := json.NewDecoder(res.Body).Decode(&ar)
				_ = res.Body.Close()
				if err != nil {
					lastErr = err
				} else if len(ar.Content) == 0 {
					lastErr = errors.New("no content returned")
				} else {
					return &models.CompletionResult{
						Content:     ar.Content[0].Text,
						UsageTokens: ar.Usage.InputTokens + ar.Usage.OutputTokens,
					}, nil
				}
			} else {
				var errResp struct {
					Error struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				_ = json.NewDecoder(res.Body).Decode(&errResp)
				_ = res.Body.Close()
				lastErr = errors.New(res.Status + ": " + errResp.Error.Message)
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}
