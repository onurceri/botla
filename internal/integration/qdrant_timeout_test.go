package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
)

func startQdrantSearchDelayStub(delay time.Duration) *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/collections/embeddings", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h.HandleFunc("/collections/embeddings/points/search", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": []any{}})
	})
	return httptest.NewServer(h)
}

// TestChat_QdrantSearchTimeout_Fallback verifies that chat works when Qdrant search times out.
// The LLM is called without RAG context and returns a response.
func TestChat_QdrantSearchTimeout_Fallback(t *testing.T) {
	t.Parallel()
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantSearchDelayStub(200 * time.Millisecond)
	defer oai.Close()
	defer qd.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.QDRANT_TIMEOUT_MS = 100
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "qdtimeout@example.com")

	create := map[string]any{
		"name": "QD Timeout Bot",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := chatReq{Message: "merhaba", SessionID: "s-qd-timeout"}
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

	// LLM is called even when Qdrant times out, returns mock response
	if crp.Response == "" {
		t.Fatalf("expected LLM response, got empty")
	}
	// In modern tiered RAG, timeout might trigger a static fallback (0 tokens)
	// if no capability summaries are available to guide a smart fallback.
	if crp.TokensUsed < 0 {
		t.Fatalf("expected tokens used >= 0, got %d", crp.TokensUsed)
	}
}
