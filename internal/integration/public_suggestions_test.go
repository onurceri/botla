package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/mock"
)

func TestPublicChatbotConfig_IncludesSuggestions(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup env: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	mockVC := &rag.MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)
	mockLLM := &rag.MockFullClient{}

	mux, q, rl, wp, _, _ := fixtures.NewTestMux(te.Cfg, te.DB, mockVC, mockLLM, mockVC)
	if q != nil {
		defer q.Stop()
	}
	if rl != nil {
		defer rl.Close()
	}
	if wp != nil {
		defer wp.Shutdown(1 * time.Second)
	}

	// Create user and bot using fixture
	ctx := testdb.CreateChatbot(t, te.DB, testdb.ChatbotFixture{
		Name:               "TestBot",
		SystemPrompt:       "prompt",
		LanguageCode:       "en",
		Model:              "gpt-3.5-turbo",
		Temperature:        0.1,
		MaxTokens:          128,
		WelcomeMessage:     "hi",
		SuggestedQuestions: []string{"Q1", "Q2"},
		SuggestionsEnabled: true,
	})
	botID := ctx.Chatbot.ID

	// Read public config
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/"+botID, bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}
	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("suggested_questions")) || !bytes.Contains([]byte(body), []byte("Q1")) {
		t.Fatalf("suggested_questions missing: %s", body)
	}
}
