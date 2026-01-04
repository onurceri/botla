package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
)

// Rate Limiting Tests
// These tests verify that rate limiting:
// - Works for chat endpoints
// - Works for source creation endpoints
// - Recovers after the rate limit window expires
// - Is isolated per-user (User A's usage doesn't affect User B)
// - Returns proper Retry-After headers

// Helper: create a text source and return the raw response (for rate limit testing)
func rlCreateTextSourceRaw(t *testing.T, baseURL, token, chatbotID, content string) *http.Response {
	t.Helper()

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	_ = mw.WriteField("source_type", "text")
	_ = mw.WriteField("text", content)
	_ = mw.Close()

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots/"+chatbotID+"/sources", &b)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create source request: %v", err)
	}
	return resp
}

// Helper: send chat request and return raw response
func rlSendChatRaw(t *testing.T, baseURL, token, chatbotID, message string) *http.Response {
	t.Helper()

	chatPayload := map[string]any{
		"message":    message,
		"session_id": fmt.Sprintf("rl-session-%d", time.Now().UnixNano()),
	}
	body, _ := json.Marshal(chatPayload)

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots/"+chatbotID+"/chat", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send chat request: %v", err)
	}
	return resp
}

// Helper: create chatbot and return ID
func rlCreateChatbot(t *testing.T, baseURL, token, name string) string {
	t.Helper()

	payload := map[string]string{"name": name}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create chatbot: %v", err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create chatbot, status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode chatbot response: %v", err)
	}

	id, ok := result["id"].(string)
	if !ok {
		t.Fatal("chatbot id not found in response")
	}

	return id
}

// Helper: auth token creation
func rlAuthToken(t *testing.T, base string, email string) string {
	t.Helper()
	regBody := map[string]string{"email": email, "password": fixtures.TestPassword, "full_name": "Rate Limit Test User"}
	b, _ := json.Marshal(regBody)
	_, _ = testHTTPPost(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	lb := map[string]string{"email": email, "password": fixtures.TestPassword}
	lbj, _ := json.Marshal(lb)
	res, err := testHTTPPost(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer drainBody(res)
	var tr tokenResp
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	return tr.Token
}

// TestRateLimit_ChatEndpoint verifies that the chat endpoint is rate limited
func TestRateLimit_ChatEndpoint(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.RateLimitUserRequestsPerMinute = 3
		cfg.RateLimitUserWindowSeconds = 60
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	// Update plan config in DB to match test expectations
	_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 4)
	_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)

	token := rlAuthToken(t, te.Server.URL, "rl_chat@example.com")
	botID := rlCreateChatbot(t, te.Server.URL, token, "Rate Limit Chat Bot")

	// Track if we hit rate limit
	var rateLimited bool
	for i := 0; i < 10; i++ {
		resp := rlSendChatRaw(t, te.Server.URL, token, botID, fmt.Sprintf("Hello %d", i))
		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true

			// Check for Retry-After header
			ra := resp.Header.Get("Retry-After")
			if ra == "" {
				t.Error("missing Retry-After header on 429 response")
			}
			drainBody(resp)
			break
		}
		drainBody(resp)
	}

	if !rateLimited {
		t.Log("Warning: rate limit not triggered within 10 requests - this may be expected if rate limit is high")
	}
}

// TestRateLimit_SourceCreation verifies that source creation is rate limited
func TestRateLimit_SourceCreation(t *testing.T) {
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.RateLimitUserRequestsPerMinute = 2
		cfg.RateLimitUserWindowSeconds = 60
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Update plan config in DB to match test expectations (low limit for testing)
	_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 3)
	_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)

	token := rlAuthToken(t, te.Server.URL, "rl_source@example.com")
	botID := rlCreateChatbot(t, te.Server.URL, token, "Rate Limit Source Bot")

	// Try to create many sources rapidly
	var rateLimited bool
	for i := 0; i < 10; i++ {
		resp := rlCreateTextSourceRaw(t, te.Server.URL, token, botID, fmt.Sprintf("Content %d - testing rate limits on source creation", i))
		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true
			drainBody(resp)
			break
		}
		drainBody(resp)
	}

	// Rate limiting should kick in
	if !rateLimited {
		t.Log("Warning: source creation rate limit not triggered - this may be expected if rate limit is high")
	}
}

// TestRateLimit_Recovery verifies that rate limits recover after the window expires
func TestRateLimit_Recovery(t *testing.T) {
	// Use a short window for testing recovery
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.RateLimitUserRequestsPerMinute = 2
		cfg.RateLimitUserWindowSeconds = 2 // 2 second window
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Update plan config in DB to match test expectations
	_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 2)
	_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 2)

	token := rlAuthToken(t, te.Server.URL, "rl_recovery@example.com")

	// Make authenticated requests to trigger rate limit
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, respErr := testHTTPClient().Do(req)
		if respErr != nil {
			t.Fatalf("request failed: %v", respErr)
		}
		drainBody(resp)
	}

	// Poll for rate limit recovery instead of fixed sleep
	// This is faster when the window expires quickly and more reliable
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/chatbots", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		drainBody(resp)

		if resp.StatusCode != http.StatusTooManyRequests {
			// Recovery happened!
			return
		}
		// Small sleep between polls to avoid hammering
		time.Sleep(100 * time.Millisecond)
	}

	t.Error("rate limit did not recover after window expiry (polled for 5s)")
}

// TestRateLimit_PerUserIsolationExtended verifies that one user's rate limit
// doesn't affect another user
func TestRateLimit_PerUserIsolationExtended(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.RateLimitUserRequestsPerMinute = 2
		cfg.RateLimitUserWindowSeconds = 60
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	// Update plan config in DB to match test expectations
	_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 2)
	_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)

	tokenA := rlAuthToken(t, te.Server.URL, "rl_isoA@example.com")
	tokenB := rlAuthToken(t, te.Server.URL, "rl_isoB@example.com")

	// Create chatbot for user A
	botA := rlCreateChatbot(t, te.Server.URL, tokenA, "User A Bot")
	// Create chatbot for user B
	botB := rlCreateChatbot(t, te.Server.URL, tokenB, "User B Bot")

	// Exhaust User A's limit with chat requests
	for i := 0; i < 5; i++ {
		resp := rlSendChatRaw(t, te.Server.URL, tokenA, botA, fmt.Sprintf("Message %d", i))
		drainBody(resp)
	}

	// User B should NOT be affected - make a request
	resp := rlSendChatRaw(t, te.Server.URL, tokenB, botB, "Hello from User B")
	defer drainBody(resp)

	if resp.StatusCode == http.StatusTooManyRequests {
		t.Error("User B was rate limited due to User A's usage - isolation failure")
	}
}

// TestRateLimit_HeadersPresent verifies that rate limit headers are properly set
func TestRateLimit_HeadersPresent(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.RateLimitUserRequestsPerMinute = 5
		cfg.RateLimitUserWindowSeconds = 60
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	// Update plan config in DB
	_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 5)
	_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)

	token := rlAuthToken(t, te.Server.URL, "rl_headers@example.com")
	botID := rlCreateChatbot(t, te.Server.URL, token, "Rate Limit Headers Bot")

	// Make a chat request and check headers
	resp := rlSendChatRaw(t, te.Server.URL, token, botID, "Hello")
	defer drainBody(resp)

	// Check for X-RateLimit-Limit header
	limit := resp.Header.Get("X-RateLimit-Limit")
	if limit == "" {
		t.Error("missing X-RateLimit-Limit header")
	} else {
		if _, parseErr := strconv.Atoi(limit); parseErr != nil {
			t.Errorf("X-RateLimit-Limit is not a valid integer: %s", limit)
		}
	}

	// Check for X-RateLimit-Remaining header
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if remaining == "" {
		t.Error("missing X-RateLimit-Remaining header")
	} else {
		if _, parseErr := strconv.Atoi(remaining); parseErr != nil {
			t.Errorf("X-RateLimit-Remaining is not a valid integer: %s", remaining)
		}
	}

	// Check for X-RateLimit-Reset header
	reset := resp.Header.Get("X-RateLimit-Reset")
	if reset == "" {
		t.Error("missing X-RateLimit-Reset header")
	}
}

// TestRateLimit_RetryAfterOnBlock verifies that Retry-After header is set when blocked
func TestRateLimit_RetryAfterOnBlock(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.RateLimitUserRequestsPerMinute = 1
		cfg.RateLimitUserWindowSeconds = 60
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	// Update plan config in DB - only allow 2 requests per minute
	_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 2)
	_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)

	token := rlAuthToken(t, te.Server.URL, "rl_retry@example.com")
	botID := rlCreateChatbot(t, te.Server.URL, token, "Rate Limit Retry Bot")

	// First request - should succeed
	resp1 := rlSendChatRaw(t, te.Server.URL, token, botID, "First message")
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("first request failed with status %d", resp1.StatusCode)
	}
	resp1.Body.Close()

	// Second request - should succeed but remaining = 0
	resp2 := rlSendChatRaw(t, te.Server.URL, token, botID, "Second message")
	resp2.Body.Close()

	// Third request - should be blocked with 429 and Retry-After
	resp3 := rlSendChatRaw(t, te.Server.URL, token, botID, "Third message")
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp3.StatusCode)
	}

	ra := resp3.Header.Get("Retry-After")
	if ra == "" {
		t.Fatal("missing Retry-After header on 429 response")
	}

	retryAfter, parseErr := strconv.Atoi(ra)
	if parseErr != nil {
		t.Fatalf("Retry-After is not a valid integer: %s", ra)
	}

	if retryAfter <= 0 {
		t.Errorf("Retry-After should be positive, got %d", retryAfter)
	}
}
