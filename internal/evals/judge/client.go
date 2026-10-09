package judge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/onurceri/botla-app/internal/evals/usage"
)

// ChatClient is the minimal interface the judge needs from a model provider.
type ChatClient interface {
	Complete(ctx context.Context, system, user string) (string, error)
	// Fingerprint identifies the model and sampling settings for cache keys.
	Fingerprint() string
}

// HTTPClient calls any OpenAI-compatible /chat/completions endpoint (OpenAI,
// OpenRouter, a local server). The judge intentionally does not reuse the
// product's LLM client: the product client wraps every prompt in its own
// "Context/Question" template, which would contaminate the rubric prompt.
type HTTPClient struct {
	BaseURL     string // e.g. https://api.openai.com/v1 or https://openrouter.ai/api/v1
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
	// JSONMode requests response_format=json_object. Leave it on unless the
	// provider/model rejects the parameter.
	JSONMode bool
	HTTP     *http.Client
	// Retries on 429/5xx/transport errors (default 3).
	Retries int
	// Backoff is the base delay between retries (default 500ms, doubled each try).
	Backoff time.Duration
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string            `json:"model"`
	Messages       []chatMessage     `json:"messages"`
	Temperature    float64           `json:"temperature"`
	MaxTokens      int               `json:"max_tokens,omitempty"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
	Usage          map[string]bool   `json:"usage,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Fingerprint implements ChatClient.
func (c *HTTPClient) Fingerprint() string {
	return fmt.Sprintf("%s|t=%.2f|max=%d|json=%t", c.Model, c.Temperature, c.MaxTokens, c.JSONMode)
}

// Complete implements ChatClient.
func (c *HTTPClient) Complete(ctx context.Context, system, user string) (string, error) {
	reqBody := chatRequest{
		Model:       c.Model,
		Messages:    []chatMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
		Temperature: c.Temperature,
		MaxTokens:   c.MaxTokens,
	}
	if c.JSONMode {
		reqBody.ResponseFormat = map[string]string{"type": "json_object"}
	}
	if strings.Contains(c.BaseURL, "openrouter.ai") {
		reqBody.Usage = map[string]bool{"include": true} // ask OpenRouter to report billed cost
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("encode judge request: %w", err)
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	retries := c.Retries
	if retries <= 0 {
		retries = 3
	}
	backoff := c.Backoff
	if backoff <= 0 {
		backoff = 500 * time.Millisecond
	}
	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"

	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("judge request cancelled: %w", ctx.Err())
			case <-time.After(backoff << (attempt - 1)):
			}
		}
		content, retryable, callErr := c.do(ctx, httpClient, url, payload)
		if callErr == nil {
			return content, nil
		}
		lastErr = callErr
		if !retryable {
			break
		}
	}
	return "", lastErr
}

func (c *HTTPClient) do(ctx context.Context, hc *http.Client, url string, payload []byte) (content string, retryable bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", false, fmt.Errorf("build judge request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	res, err := hc.Do(req)
	if err != nil {
		// Context cancellation and budget exhaustion are not worth retrying.
		retry := ctx.Err() == nil && !errors.Is(err, usage.ErrBudgetExceeded)
		return "", retry, fmt.Errorf("judge request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return "", true, fmt.Errorf("read judge response: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		retry := res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500
		return "", retry, fmt.Errorf("judge http %d: %s", res.StatusCode, truncate(string(body), 300))
	}
	var cr chatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", true, fmt.Errorf("decode judge response: %w", err)
	}
	if len(cr.Choices) == 0 || strings.TrimSpace(cr.Choices[0].Message.Content) == "" {
		return "", true, errors.New("judge returned an empty completion")
	}
	return cr.Choices[0].Message.Content, false, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
