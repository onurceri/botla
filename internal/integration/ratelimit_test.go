package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
)

func TestRateLimit_Chat_Sources(t *testing.T) {
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
	token := authToken(t, te.Server.URL, "rl@example.com")

	// create bot
	create := map[string]any{"name": "RL Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// first three chat requests allowed
	cr := chatReq{Message: "merhaba", SessionID: "s3"}
	crb, _ := json.Marshal(cr)
	for i := 0; i < 3; i++ {
		reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
		reqCh.Header.Set("Authorization", "Bearer "+token)
		reqCh.Header.Set("Content-Type", "application/json")
		resCh, _ := testHTTPClient().Do(reqCh)
		if resCh.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resCh.StatusCode)
		}
		resCh.Body.Close()
	}
	// fourth chat should be 429
	reqCh4, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh4.Header.Set("Authorization", "Bearer "+token)
	reqCh4.Header.Set("Content-Type", "application/json")
	resCh4, _ := testHTTPClient().Do(reqCh4)
	if resCh4.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resCh4.StatusCode)
	}
	resCh4.Body.Close()
}
