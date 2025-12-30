package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BaseClient struct {
	baseURL    string
	apiKey     string
	headers    map[string]string
	httpClient *http.Client
}

func NewBaseClient(baseURL, apiKey string, headers map[string]string) *BaseClient {
	return &BaseClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		headers:    headers,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func NewBaseClientWithHTTPClient(baseURL, apiKey string, headers map[string]string, httpClient *http.Client) *BaseClient {
	return &BaseClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		headers:    headers,
		httpClient: httpClient,
	}
}

func (c *BaseClient) Post(ctx context.Context, path string, body, responseTarget interface{}) error {
	var lastErr error

	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
		}

		lastErr = c.doRequest(ctx, path, body, responseTarget)
		if lastErr == nil {
			return nil
		}

		if !shouldRetry(lastErr) {
			return lastErr
		}
	}

	return lastErr
}

func (c *BaseClient) doRequest(ctx context.Context, path string, body, responseTarget interface{}) error {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", res.Status)
	}

	if responseTarget != nil {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if len(resBody) == 0 {
			return fmt.Errorf("empty response body")
		}

		if err := json.Unmarshal(resBody, responseTarget); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	errStr := err.Error()
	if len(errStr) < 3 {
		return false
	}

	switch {
	case errStr == "429 Too Many Requests":
		return true
	case errStr[:3] == "500":
		return true
	case errStr[:3] == "502":
		return true
	case errStr[:3] == "503":
		return true
	case errStr[:3] == "504":
		return true
	default:
		return false
	}
}
