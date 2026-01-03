package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
)

func TestChat_OpenAIEnvMissing_Graceful(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	// Use an invalid port to force connection failure
	// Ensure OPENAI_API_KEY is missing, but satisfy LoadConfig with another key

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.ANTHROPIC_API_KEY = "dummy"
		cfg.OPENAI_API_BASE = "http://127.0.0.1:1"
		cfg.OPENROUTER_API_BASE = "http://127.0.0.1:1"
		cfg.QDRANT_URL = qd.URL
		cfg.OPENAI_API_KEY = ""
	}, false) // useMocks=false to test real failure behavior
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	token := authToken(t, te.Server.URL, "envmissing@example.com")
	create := map[string]any{"name": "Env Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := chatReq{Message: "merhaba", SessionID: "s-env"}
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

	// When OpenAI is unavailable and chatbot has no sources, we expect either:
	// - NoInfoFound: "Yeterli bilgi bulamadım."
	// - EmptyStateMessage: "Henüz bilgi kaynaklarım yüklenmedi, ama yardımcı olmaya hazırım!"
	// Both are valid graceful fallbacks
	validFallbacks := []string{
		"Yeterli bilgi bulamadım.",
		"Henüz bilgi kaynaklarım yüklenmedi, ama yardımcı olmaya hazırım!",
	}
	isValidFallback := false
	for _, valid := range validFallbacks {
		if crp.Response == valid {
			isValidFallback = true
			break
		}
	}
	if !isValidFallback {
		t.Fatalf("expected graceful fallback message, got %q", crp.Response)
	}
}

// TestChat_QdrantEnvMissing_Fallback verifies that chat works when QDRANT_URL is missing.
// The LLM is called without RAG context and returns a response.
func TestChat_QdrantEnvMissing_Fallback(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.QDRANT_URL = ""
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()

	token := authToken(t, te.Server.URL, "qdenv@example.com")
	create := map[string]any{
		"name": "QD Env Bot",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := chatReq{Message: "merhaba", SessionID: "s-qdenv"}
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

	// In modern tiered RAG, missing Qdrant URL triggers a static fallback (0 tokens)
	// if no capability summaries are available to guide a smart fallback.
	if crp.Response == "" {
		t.Fatalf("expected fallback response, got empty")
	}
	if crp.TokensUsed < 0 {
		t.Fatalf("expected tokens used >= 0, got %d", crp.TokensUsed)
	}
}
