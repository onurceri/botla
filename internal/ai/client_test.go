package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBaseClient_SuccessfulRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header, got %s", r.Header.Get("Content-Type"))
		}

		var reqBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	client := NewBaseClient(server.URL, "test-key", map[string]string{"X-Custom": "custom-value"})
	ctx := context.Background()

	reqBody := map[string]string{"test": "data"}
	var respBody map[string]string

	err := client.Post(ctx, "/test", reqBody, &respBody)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if respBody["result"] != "success" {
		t.Errorf("Expected success response, got %v", respBody)
	}
}

func TestBaseClient_RateLimitRetry(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	client := NewBaseClient(server.URL, "test-key", nil)
	ctx := context.Background()

	reqBody := map[string]string{"test": "data"}
	var respBody map[string]string

	startTime := time.Now()
	err := client.Post(ctx, "/test", reqBody, &respBody)
	duration := time.Since(startTime)

	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
	if requestCount != 3 {
		t.Errorf("Expected 3 requests (2 rate limits + 1 success), got %d", requestCount)
	}
	if respBody["result"] != "success" {
		t.Errorf("Expected success response, got %v", respBody)
	}
	if duration < 400*time.Millisecond {
		t.Errorf("Expected at least 400ms delay (200ms + 400ms), got %v", duration)
	}
}

func TestBaseClient_ServerErrorRetry(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	client := NewBaseClient(server.URL, "test-key", nil)
	ctx := context.Background()

	reqBody := map[string]string{"test": "data"}
	var respBody map[string]string

	err := client.Post(ctx, "/test", reqBody, &respBody)

	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
	if requestCount != 2 {
		t.Errorf("Expected 2 requests (1 error + 1 success), got %d", requestCount)
	}
}

func TestBaseClient_NonRetryableError(t *testing.T) {
	testCases := []struct {
		name          string
		statusCode    int
		expectedError string
	}{
		{"401 Unauthorized", http.StatusUnauthorized, "401 Unauthorized"},
		{"400 Bad Request", http.StatusBadRequest, "400 Bad Request"},
		{"404 Not Found", http.StatusNotFound, "404 Not Found"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			client := NewBaseClient(server.URL, "test-key", nil)
			ctx := context.Background()

			reqBody := map[string]string{"test": "data"}
			var respBody map[string]string

			err := client.Post(ctx, "/test", reqBody, &respBody)

			if err == nil {
				t.Error("Expected error for non-retryable status code")
			}
			if requestCount != 1 {
				t.Errorf("Expected 1 request (no retries), got %d", requestCount)
			}
		})
	}
}

func TestBaseClient_ContextCancellation(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	client := NewBaseClient(server.URL, "test-key", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	reqBody := map[string]string{"test": "data"}
	var respBody map[string]string

	err := client.Post(ctx, "/test", reqBody, &respBody)

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context error, got %v", err)
	}
}

func TestBaseClient_ExhaustedRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewBaseClient(server.URL, "test-key", nil)
	ctx := context.Background()

	reqBody := map[string]string{"test": "data"}
	var respBody map[string]string

	err := client.Post(ctx, "/test", reqBody, &respBody)

	if err == nil {
		t.Error("Expected error after exhausting retries")
	}
}

func TestBaseClient_MalformedJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid json"))
	}))
	defer server.Close()

	client := NewBaseClient(server.URL, "test-key", nil)
	ctx := context.Background()

	reqBody := map[string]string{"test": "data"}
	var respBody map[string]string

	err := client.Post(ctx, "/test", reqBody, &respBody)

	if err == nil {
		t.Error("Expected error for malformed JSON")
	}
}

func TestBaseClient_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewBaseClient(server.URL, "test-key", nil)
	ctx := context.Background()

	reqBody := map[string]string{"test": "data"}
	var respBody map[string]string

	err := client.Post(ctx, "/test", reqBody, &respBody)

	if err == nil {
		t.Error("Expected error for empty body")
	}
}

func TestBaseClient_NetworkTimeoutDuringRetries(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 100 * time.Millisecond}
	client := NewBaseClientWithHTTPClient(server.URL, "test-key", nil, httpClient)
	ctx := context.Background()

	reqBody := map[string]string{"test": "data"}
	var respBody map[string]string

	err := client.Post(ctx, "/test", reqBody, &respBody)

	if err == nil {
		t.Error("Expected error due to network timeout")
	}
	if requestCount < 2 {
		t.Errorf("Expected at least 2 retry attempts, got %d", requestCount)
	}
}

func TestBaseClient_CustomHeaders(t *testing.T) {
	receivedHeaders := make(map[string]string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders["HTTP-Referer"] = r.Header.Get("HTTP-Referer")
		receivedHeaders["X-Title"] = r.Header.Get("X-Title")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	customHeaders := map[string]string{
		"HTTP-Referer": "https://botla.app",
		"X-Title":      "Botla",
	}
	client := NewBaseClient(server.URL, "test-key", customHeaders)
	ctx := context.Background()

	reqBody := map[string]string{"test": "data"}
	var respBody map[string]string

	err := client.Post(ctx, "/test", reqBody, &respBody)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if receivedHeaders["HTTP-Referer"] != "https://botla.app" {
		t.Errorf("Expected custom header HTTP-Referer, got %s", receivedHeaders["HTTP-Referer"])
	}
	if receivedHeaders["X-Title"] != "Botla" {
		t.Errorf("Expected custom header X-Title, got %s", receivedHeaders["X-Title"])
	}
}
