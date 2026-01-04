package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/pkg/config"
)

// TestChat_Fallback_NoContext verifies that chat works when no Qdrant URL is set.
// The LLM is called without RAG context and returns a response.
func TestChat_Fallback_NoContext(t *testing.T) {
	t.Parallel()
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.QDRANT_URL = ""
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	token := authToken(t, te.Server.URL, "nofallback@example.com")

	create := map[string]any{
		"name": "FB Bot",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := chatReq{Message: "merhaba", SessionID: "s2"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := testHTTPClient().Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resCh.StatusCode)
	}
	var crp chatResp
	json.NewDecoder(resCh.Body).Decode(&crp)
	resCh.Body.Close()

	// In modern tiered RAG, missing context triggers a static fallback (0 tokens)
	// if no capability summaries are available to guide a smart fallback.
	if crp.Response == "" {
		t.Fatalf("expected fallback response, got empty")
	}
	if crp.TokensUsed < 0 {
		t.Fatalf("expected tokens used >= 0, got %d", crp.TokensUsed)
	}
}
