package integration

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/stretchr/testify/mock"
)

func TestPublicChatbotConfig_IncludesSuggestions(t *testing.T) {
	te, _ := SetupTestEnv()
	defer TeardownTestEnv(te)

	mockVC := &rag.MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)
	mockLLM := &rag.MockFullClient{}

	mux := NewTestMux(te.Cfg, te.DB, nil, mockLLM, mockVC)

	// Create user and bot
	userID := createTestUser(t, te.DB)
	bot := &models.Chatbot{
		UserID:               userID,
		Name:                 "TestBot",
		SystemPrompt:         "prompt",
		LanguageCode:         "en",
		Model:                "gpt-3.5-turbo",
		Temperature:          0.1,
		MaxTokens:            128,
		ThemeColor:           "#000000",
		WelcomeMessage:       "hi",
		Position:             "bottom-right",
		BotMessageColor:      "#000000",
		UserMessageColor:     "#ffffff",
		BotMessageTextColor:  "#ffffff",
		UserMessageTextColor: "#000000",
		ChatFontFamily:       "Inter",
		ChatHeaderColor:      "#000000",
		ChatHeaderTextColor:  "#ffffff",
		ChatBackgroundColor:  "#ffffff",
		SuggestedQuestions:   []string{"Q1", "Q2"},
		SuggestionsEnabled:   true,
	}
	botID, err := db.CreateChatbot(context.Background(), te.DB, bot)
	if err != nil {
		t.Fatalf("create bot: %v", err)
	}

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

// helpers
func createTestUser(t *testing.T, db *sql.DB) string {
	var id string
	email := "test+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	var freePlanID string
	if err := db.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&id); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}
