package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRateLimit_PerUserIsolation(t *testing.T) {
	t.Setenv("RATE_LIMIT_USER_REQUESTS_PER_MINUTE", "1")
	t.Setenv("RATE_LIMIT_USER_WINDOW_SECONDS", "60")
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

	tokenA := authToken(t, te.Server.URL, "isoA@example.com")
	tokenB := authToken(t, te.Server.URL, "isoB@example.com")

	// create bot for A
	createA := map[string]any{"name": "ISO A Bot"}
	cbjA, _ := json.Marshal(createA)
	reqCA, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbjA))
	reqCA.Header.Set("Authorization", "Bearer "+tokenA)
	reqCA.Header.Set("Content-Type", "application/json")
	resCA, _ := http.DefaultClient.Do(reqCA)
	var botA chatbot
	json.NewDecoder(resCA.Body).Decode(&botA)
	resCA.Body.Close()

	// create bot for B
	createB := map[string]any{"name": "ISO B Bot"}
	cbjB, _ := json.Marshal(createB)
	reqCB, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbjB))
	reqCB.Header.Set("Authorization", "Bearer "+tokenB)
	reqCB.Header.Set("Content-Type", "application/json")
	resCB, _ := http.DefaultClient.Do(reqCB)
	var botB chatbot
	json.NewDecoder(resCB.Body).Decode(&botB)
	resCB.Body.Close()

	// first chat for A allowed
	cr := chatReq{Message: "merhaba", SessionID: "s-iso"}
	crb, _ := json.Marshal(cr)
	reqA1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+botA.ID+"/chat", bytes.NewReader(crb))
	reqA1.Header.Set("Authorization", "Bearer "+tokenA)
	reqA1.Header.Set("Content-Type", "application/json")
	resA1, _ := http.DefaultClient.Do(reqA1)
	if resA1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resA1.StatusCode)
	}
	resA1.Body.Close()

	// first chat for B should also be allowed (separate user key)
	reqB1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+botB.ID+"/chat", bytes.NewReader(crb))
	reqB1.Header.Set("Authorization", "Bearer "+tokenB)
	reqB1.Header.Set("Content-Type", "application/json")
	resB1, _ := http.DefaultClient.Do(reqB1)
	if resB1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resB1.StatusCode)
	}
	resB1.Body.Close()

	// second chat for B should be blocked due to per-user limit 1
	reqB2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+botB.ID+"/chat", bytes.NewReader(crb))
	reqB2.Header.Set("Authorization", "Bearer "+tokenB)
	reqB2.Header.Set("Content-Type", "application/json")
	resB2, _ := http.DefaultClient.Do(reqB2)
	if resB2.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resB2.StatusCode)
	}
	resB2.Body.Close()
}
