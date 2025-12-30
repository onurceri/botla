package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestAnalytics_UpdatesAfterChat(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	token := authToken(t, te.Server.URL, "anupd@example.com")

	// create chatbot
	create := map[string]any{"name": "AN Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// perform chat (user + assistant messages)
	cr := chatReq{Message: "merhaba", SessionID: "s5"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := http.DefaultClient.Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resCh.StatusCode)
	}
	resCh.Body.Close()

	// read analytics and assert totals increased
	reqA, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/analytics", nil)
	reqA.Header.Set("Authorization", "Bearer "+token)
	resA, _ := http.DefaultClient.Do(reqA)
	if resA.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resA.StatusCode)
	}
	var series []struct {
		Date          string
		Messages      int
		Conversations int
	}
	json.NewDecoder(resA.Body).Decode(&series)
	resA.Body.Close()
	var totalM, totalC int
	for _, p := range series {
		totalM += p.Messages
		totalC += p.Conversations
	}
	if totalM < 2 || totalC < 1 {
		t.Fatalf("expected totals >= (2 messages, 1 conv), got (%d, %d)", totalM, totalC)
	}
}
