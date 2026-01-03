package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/stretchr/testify/mock"
)


func TestChat_OpenAIError_GracefulMessage(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Replace the real factory/client with a mock that always fails
	mockLLM := &rag.MockFullClient{}
	mockLLM.On("CreateEmbedding", mock.Anything, mock.Anything).Return([]float32{0.1}, nil)
	mockLLM.On("CreateCompletionWithTools", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("simulated openai error"))

	mockVC := &rag.MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)
	mockVC.On("SearchSimilar", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]rag.SearchResult{
		{
			ID:    "1",
			Score: 0.9,
			Payload: rag.EmbeddingPayload{
				OriginalText: "some context",
				SourceID:     "00000000-0000-0000-0000-000000000001",
				SourceType:   "file",
			},
		},
	}, nil)

	// Actually, let's just create a custom mux for this test
	mux, q, rl, wp, _, _ := fixtures.NewTestMux(te.Cfg, te.DB, te.VectorStore, mockLLM, mockVC)
	if q != nil {
		defer q.Stop()
	}
	if rl != nil {
		defer rl.Close()
	}
	if wp != nil {
		defer wp.Shutdown(1 * time.Second)
	}
	ts := httptest.NewServer(mux)
	defer ts.Close()

	token := authToken(t, ts.URL, "chaterr@example.com")

	// create chatbot
	create := map[string]any{"name": "Chat Err Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// chat
	cr := chatReq{Message: "merhaba", SessionID: "s3"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := testHTTPClient().Do(reqCh)
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
