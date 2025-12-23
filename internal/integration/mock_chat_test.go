package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onurceri/botla-co/internal/rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockChatFlow(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	mockLLM := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)

	// Setup Mux with mocks
	mux := NewTestMux(te.Cfg, te.DB, te.VectorStore, mockLLM, mockVC)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 1. Setup Mock Expectations

	// Query embedding mock
	mockLLM.On("CreateEmbedding", mock.Anything, "Hello bot").Return([]float32{0.1, 0.2}, nil).Once()

	// Vector Search mock
	mockVC.On("SearchSimilar", mock.Anything, []float32{0.1, 0.2}, mock.Anything, 3).Return([]rag.SearchResult{
		{
			ID:    "chunk-1",
			Score: 0.9,
			Payload: rag.EmbeddingPayload{
				OriginalText: "This is some mock knowledge context.",
				SourceID:     "src-1",
				SourceType:   "text",
			},
		},
	}, nil).Once()

	// LLM Completion mock
	mockLLM.On("CreateCompletionWithTools", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&rag.ChatResponseWithTools{
		Choices: []struct {
			Message      rag.ChatMessage `json:"message"`
			FinishReason string          `json:"finish_reason"`
		}{{
			Message: rag.ChatMessage{Role: "assistant", Content: strPtr("I am a mock assistant answering based on context.")},
		}},
		Usage: struct {
			TotalTokens int `json:"total_tokens"`
		}{TotalTokens: 50},
	}, nil).Once()

	token := authToken(t, ts.URL, "mockchat@example.com")

	// 2. Create Chatbot
	create := map[string]any{"name": "Mock Chat Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 3. Chat
	chatReq := map[string]string{
		"message":    "Hello bot",
		"session_id": "s-mock-1",
	}
	crb, _ := json.Marshal(chatReq)
	reqCh, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chatbots/"+bot.ID+"/chat", strings.NewReader(string(crb)))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := http.DefaultClient.Do(reqCh)

	assert.Equal(t, http.StatusOK, resCh.StatusCode)

	var chatResp map[string]any
	json.NewDecoder(resCh.Body).Decode(&chatResp)
	resCh.Body.Close()

	assert.Equal(t, "I am a mock assistant answering based on context.", chatResp["response"])
	assert.Equal(t, float64(50), chatResp["tokens_used"])

	mockLLM.AssertExpectations(t)
	mockVC.AssertExpectations(t)
}
