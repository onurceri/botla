package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"database/sql"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/testdb"
)

func TestPublicChatbotConfig_SuggestionsCacheKeyedByUpdatedAt(t *testing.T) {
	pool := mustInitDB(t)
	chatbotRepo := repository.NewPostgresChatbotRepo(pool)
	var uid string
	var freePlanID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := "cache_" + time.Now().Format("150405.000000") + "@test"
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}
	bot := &models.Chatbot{
		UserID:               uid,
		Name:                 "Bot",
		SystemPrompt:         "p",
		LanguageCode:         "en",
		Model:                "gpt-3.5-turbo",
		Temperature:          0.1,
		MaxTokens:            64,
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
		SuggestedQuestions:   []string{"A", "B"},
		SuggestionsEnabled:   true,
	}
	bid, err := repository.NewPostgresChatbotRepo(pool).Create(context.Background(), bot)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/"+bid, nil)
	w1 := httptest.NewRecorder()
	PublicChatbotConfig(chatbotRepo)(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("status1: %d", w1.Code)
	}
	var m1 map[string]any
	_ = json.Unmarshal(w1.Body.Bytes(), &m1)
	if len(m1["suggested_questions"].([]any)) != 2 {
		t.Fatalf("len1")
	}

	time.Sleep(10 * time.Millisecond)
	if _, err := pool.Exec(`UPDATE chatbots SET suggested_questions=$1, updated_at=NOW() WHERE id=$2`, jsonArr([]string{"C"}), bid); err != nil {
		t.Fatalf("upd: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/"+bid, nil)
	w2 := httptest.NewRecorder()
	PublicChatbotConfig(chatbotRepo)(w2, req2)
	var m2 map[string]any
	_ = json.Unmarshal(w2.Body.Bytes(), &m2)
	if len(m2["suggested_questions"].([]any)) != 1 {
		t.Fatalf("len2")
	}
}

func jsonArr(in []string) []byte { b, _ := json.Marshal(in); return b }

func mustInitDB(t *testing.T) *sql.DB {
	t.Helper()
	return testdb.OpenTestDB(t)
}
