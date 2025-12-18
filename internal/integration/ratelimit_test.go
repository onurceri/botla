package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRateLimit_Chat_Sources(t *testing.T) {
	t.Setenv("RATE_LIMIT_USER_REQUESTS_PER_MINUTE", "3")
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
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{rate_limits}', '{"requests_per_minute": 4, "window_seconds": 60}'::jsonb) WHERE code = 'free'`)
	if err != nil {
		t.Fatalf("failed to update rate limits: %v", err)
	}
	token := authToken(t, te.Server.URL, "rl@example.com")

	// create bot
	create := map[string]any{"name": "RL Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
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
		resCh, _ := http.DefaultClient.Do(reqCh)
		if resCh.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resCh.StatusCode)
		}
		resCh.Body.Close()
	}
	// fourth chat should be 429
	reqCh4, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh4.Header.Set("Authorization", "Bearer "+token)
	reqCh4.Header.Set("Content-Type", "application/json")
	resCh4, _ := http.DefaultClient.Do(reqCh4)
	if resCh4.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resCh4.StatusCode)
	}
	resCh4.Body.Close()
}
