package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/rag"
	"github.com/stretchr/testify/mock"
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
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Replace the real factory/client with a mock that always fails
	mockLLM := &rag.MockFullClient{}
	mockLLM.On("CreateEmbedding", mock.Anything, mock.Anything).Return([]float32{0.1}, nil)
	mockLLM.On("CreateCompletionWithTools", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("simulated openai error"))

	mockVC := &rag.MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)
	mockVC.On("SearchSimilar", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]rag.SearchResult{}, nil)

	// Actually, let's just create a custom mux for this test
	mux := NewTestMux(te.Cfg, te.DB, te.VectorStore, mockLLM, mockVC)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	token := authToken(t, ts.URL, "chaterr@example.com")

	// create chatbot
	create := map[string]any{"name": "Chat Err Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// chat
	cr := chatReq{Message: "merhaba", SessionID: "s3"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
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
