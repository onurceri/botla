package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/pkg/config"
)

func TestRateLimit_Headers(t *testing.T) {
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
	_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 3)
	_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)

	token := authToken(t, te.Server.URL, "rlhdr@example.com")

	// create bot
	create := map[string]any{"name": "RL HDR Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// first request
	cr := chatReq{Message: "merhaba", SessionID: "s-hdr"}
	crb, _ := json.Marshal(cr)
	req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	res1, _ := testHTTPClient().Do(req1)
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
	res2, _ := testHTTPClient().Do(req2)
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
	res3, _ := testHTTPClient().Do(req3)
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
