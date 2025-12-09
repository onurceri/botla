package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

type GoogleAIClient struct {
	apiKey       string
	http         *http.Client
	base         string
	defaultModel string
}

func NewGoogleAIClientFromEnv() (*GoogleAIClient, error) {
	k := os.Getenv("GOOGLE_AI_API_KEY")
	if k == "" {
		return nil, errors.New("GOOGLE_AI_API_KEY is empty")
	}

	b := os.Getenv("GOOGLE_AI_API_BASE")
	if b == "" {
		b = "https://generativelanguage.googleapis.com/v1beta"
	}

	return &GoogleAIClient{
		apiKey:       k,
		http:         &http.Client{Timeout: 60 * time.Second},
		base:         b,
		defaultModel: "gemini-1.5-flash",
	}, nil
}

func (c *GoogleAIClient) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{
		Name:              "gemini-1.5-flash",
		Provider:          "google",
		MaxTokens:         1000000,
		SupportedFeatures: []string{"chat"},
	}
}

// Google AI API structs
type googlePart struct {
	Text string `json:"text"`
}

type googleContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []googlePart `json:"parts"`
}

type googleRequest struct {
	Contents          []googleContent  `json:"contents"`
	SystemInstruction *googleContent   `json:"systemInstruction,omitempty"`
	GenerationConfig  *googleGenConfig `json:"generationConfig,omitempty"`
}

type googleGenConfig struct {
	Temperature     float32 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type googleResponse struct {
	Candidates []struct {
		Content struct {
			Parts []googlePart `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (c *GoogleAIClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	model := params.Model
	if model == "" {
		model = c.defaultModel
	}
	// Strip "google:" prefix if present
	if len(model) > 7 && model[:7] == "google:" {
		model = model[7:]
	}

	userContent := "Context:\n" + params.Context + "\n\nQuestion:\n" + params.UserMessage

	reqBody := googleRequest{
		Contents: []googleContent{
			{
				Role:  "user",
				Parts: []googlePart{{Text: userContent}},
			},
		},
		GenerationConfig: &googleGenConfig{
			Temperature:     params.Temperature,
			MaxOutputTokens: params.MaxTokens,
		},
	}

	if params.SystemPrompt != "" {
		reqBody.SystemInstruction = &googleContent{
			Parts: []googlePart{{Text: params.SystemPrompt}},
		}
	}

	b, _ := json.Marshal(reqBody)
	var lastErr error

	for attempt := 0; attempt < 4; attempt++ {
		url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.base, model, c.apiKey)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")

		res, err := c.http.Do(req)
		switch {
		case err != nil:
			lastErr = err
		case res.StatusCode != http.StatusOK:
			var errResp struct {
				Error struct {
					Message string `json:"message"`
					Status  string `json:"status"`
				} `json:"error"`
			}
			_ = json.NewDecoder(res.Body).Decode(&errResp)
			_ = res.Body.Close()
			lastErr = errors.New(res.Status + ": " + errResp.Error.Message)
		default:
			var gr googleResponse
			err := json.NewDecoder(res.Body).Decode(&gr)
			_ = res.Body.Close()
			switch {
			case err != nil:
				lastErr = err
			case len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0:
				lastErr = errors.New("no content returned")
			default:
				return &models.CompletionResult{
					Content:     gr.Candidates[0].Content.Parts[0].Text,
					UsageTokens: gr.UsageMetadata.TotalTokenCount,
				}, nil
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
	}
	return nil, lastErr
}
