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
)

func TestPublicChatbotConfig_IncludesSuggestions(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer TeardownTestEnv(te)
	mux := NewTestMux(te.Cfg, te.DB)

	// Create user and bot
	userID := createTestUser(t, te.DB)
	bot := &models.Chatbot{
		UserID:               userID,
		Name:                 "TestBot",
		SystemPrompt:         "prompt",
		Language:             "en",
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
	if !(bytes.Contains([]byte(body), []byte("suggested_questions")) && bytes.Contains([]byte(body), []byte("Q1"))) {
		t.Fatalf("suggested_questions missing: %s", body)
	}
}

// helpers
func createTestUser(t *testing.T, db *sql.DB) string {
	var id string
	email := "test+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	if err := db.QueryRow(`INSERT INTO users (email, password_hash) VALUES ($1,$2) RETURNING id`, email, "x").Scan(&id); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}
