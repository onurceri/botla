package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

type chatReq struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

type chatResp struct {
	Response    string           `json:"response"`
	TokensUsed  int              `json:"tokens_used"`
	SourcesUsed []map[string]any `json:"sources_used"`
}

func TestChat_StubbedContext(t *testing.T) {
	oai := NewLLMMock(t)
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

	token := authToken(t, te.Server.URL, "chat@example.com")

	// create chatbot
	create := map[string]any{"name": "Chat Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// chat
	cr := chatReq{Message: "merhaba", SessionID: "s1"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := http.DefaultClient.Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resCh.StatusCode)
	}
	var crp chatResp
	json.NewDecoder(resCh.Body).Decode(&crp)
	resCh.Body.Close()
	if crp.Response == "" || crp.TokensUsed == 0 {
		t.Fatalf("invalid chat response")
	}
}
