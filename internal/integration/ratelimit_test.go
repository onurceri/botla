package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRateLimit_Chat_Sources(t *testing.T) {
	t.Setenv("RATE_LIMIT_REQUESTS", "3")
	t.Setenv("RATE_LIMIT_WINDOW_SECONDS", "60")
	oai := startOpenAIStub()
	qd := startQdrantStub()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()
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
