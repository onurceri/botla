package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestRateLimit_Headers(t *testing.T) {
	t.Setenv("RATE_LIMIT_USER_REQUESTS_PER_MINUTE", "2")
	t.Setenv("RATE_LIMIT_USER_WINDOW_SECONDS", "60")
	oai := NewLLMMock(t)
	qd := startQdrantStub()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("QDRANT_URL", qd.URL)
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	// Update plan config in DB to match test expectations
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{rate_limits}', '{"requests_per_minute": 3, "window_seconds": 60}'::jsonb) WHERE code = 'free'`)
	if err != nil {
		t.Fatalf("failed to update rate limits: %v", err)
	}

	token := authToken(t, te.Server.URL, "rlhdr@example.com")

	// create bot
	create := map[string]any{"name": "RL HDR Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// first request
	cr := chatReq{Message: "merhaba", SessionID: "s-hdr"}
	crb, _ := json.Marshal(cr)
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res1.StatusCode)
	}
	if lim := res1.Header.Get("X-RateLimit-Limit"); lim != "3" {
		t.Fatalf("limit header expected 3, got %q", lim)
	}
	if rem := res1.Header.Get("X-RateLimit-Remaining"); rem != "1" {
		t.Fatalf("remaining expected 1, got %q", rem)
	}
	res1.Body.Close()

	// second request allowed, remaining 0
	req2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	res2, _ := http.DefaultClient.Do(req2)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res2.StatusCode)
	}
	if rem := res2.Header.Get("X-RateLimit-Remaining"); rem != "0" {
		t.Fatalf("remaining expected 0, got %q", rem)
	}
	res2.Body.Close()

	// third request blocked with 429 and Retry-After set
	req3, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req3.Header.Set("Authorization", "Bearer "+token)
	req3.Header.Set("Content-Type", "application/json")
	res3, _ := http.DefaultClient.Do(req3)
	if res3.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", res3.StatusCode)
	}
	ra := res3.Header.Get("Retry-After")
	if ra == "" {
		t.Fatalf("missing Retry-After header")
	}
	if n, err := strconv.Atoi(ra); err != nil || n <= 0 {
		t.Fatalf("invalid Retry-After: %q", ra)
	}
	res3.Body.Close()
}
