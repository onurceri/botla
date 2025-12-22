package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestChat_OpenAIEnvMissing_Graceful(t *testing.T) {
	oai := NewLLMMock(t)
	qd := startQdrantStub()
	// Use an invalid port to force connection failure
	t.Setenv("OPENAI_API_BASE", "http://127.0.0.1:1")
	t.Setenv("OPENROUTER_API_BASE", "http://127.0.0.1:1")
	t.Setenv("QDRANT_URL", qd.URL)

	// Ensure OPENAI_API_KEY is missing, but satisfy LoadConfig with another key
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "dummy")

	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	token := authToken(t, te.Server.URL, "envmissing@example.com")
	create := map[string]any{"name": "Env Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := chatReq{Message: "merhaba", SessionID: "s-env"}
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

	if crp.Response != "Şu an bir hata oluştu, lütfen tekrar deneyin." {
		t.Fatalf("expected graceful error message, got %q", crp.Response)
	}
}

// TestChat_QdrantEnvMissing_Fallback verifies that chat works when QDRANT_URL is missing.
// The LLM is called without RAG context and returns a response.
func TestChat_QdrantEnvMissing_Fallback(t *testing.T) {
	oai := NewLLMMock(t)
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("QDRANT_URL", "")
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	defer oai.Close()

	token := authToken(t, te.Server.URL, "qdenv@example.com")
	create := map[string]any{
		"name": "QD Env Bot",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := chatReq{Message: "merhaba", SessionID: "s-qdenv"}
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

	// LLM is called even when QDRANT_URL is empty, returns mock response
	if crp.Response == "" {
		t.Fatalf("expected LLM response, got empty")
	}
	if crp.TokensUsed <= 0 {
		t.Fatalf("expected tokens used > 0, got %d", crp.TokensUsed)
	}
}
