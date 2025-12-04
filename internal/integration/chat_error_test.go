package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// startOpenAIErrorStub returns embeddings OK, but chat completions 500
func startOpenAIErrorStub() *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusInternalServerError)
	})
	return httptest.NewServer(h)
}

func TestChat_OpenAIError_GracefulMessage(t *testing.T) {
	oai := startOpenAIErrorStub()
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

	token := authToken(t, te.Server.URL, "chaterr@example.com")

	// create chatbot
	create := map[string]any{"name": "Chat Err Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// chat
	cr := chatReq{Message: "merhaba", SessionID: "s3"}
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
	if crp.Response != "Şu an bir hata oluştu, lütfen tekrar deneyin." || crp.TokensUsed != 0 {
		t.Fatalf("expected graceful error message, got %q/%d", crp.Response, crp.TokensUsed)
	}
}
