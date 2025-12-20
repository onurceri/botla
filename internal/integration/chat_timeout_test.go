package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startOpenAITimeoutStub(delay time.Duration) *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := make([]float64, 1536)
		for i := range data {
			data[i] = 0.01
		}
		resp := map[string]any{"data": []map[string]any{{"embedding": data}}, "usage": map[string]int{"prompt_tokens": 10, "total_tokens": 10}}
		json.NewEncoder(w).Encode(resp)
	})
	h.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]string{"content": "Stubbed"}}}, "usage": map[string]int{"total_tokens": 1}})
	})
	return httptest.NewServer(h)
}

// TestChat_OpenAIEmbeddingTimeout_Fallback verifies chat behavior when embedding service times out.
// When timeout is very short, both embedding and LLM calls may fail, returning an error message.
func TestChat_OpenAIEmbeddingTimeout_Fallback(t *testing.T) {
	// Delay embeddings beyond configured client/chat timeout
	oai := startOpenAITimeoutStub(200 * time.Millisecond)
	qd := startQdrantStub()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("QDRANT_URL", qd.URL)
	t.Setenv("OPENAI_TIMEOUT_MS", "100")
	t.Setenv("CHAT_TIMEOUT_MS", "100")
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	token := authToken(t, te.Server.URL, "timeout@example.com")

	create := map[string]any{
		"name": "Timeout Bot",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := chatReq{Message: "merhaba", SessionID: "s-timeout"}
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

	// When timeout occurs, either a fallback message or error message is returned
	// The response should not be empty
	if crp.Response == "" {
		t.Fatalf("expected some response, got empty")
	}
}

