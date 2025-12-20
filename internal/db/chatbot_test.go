package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
)

func createUser(t *testing.T, db *sql.DB) string {
	t.Helper()
	email := fmt.Sprintf("dbu+%d@example.com", time.Now().UnixNano())
	var id string
	var freePlanID string
	if err := db.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&id); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}

func TestChatbot_CRUD_DB(t *testing.T) {
	db := testdb.OpenTestDB(t)
	uid := createUser(t, db)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, uid)
	})
	b := &models.Chatbot{
		UserID:               uid,
		Name:                 "DB Bot",
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
		SuggestedQuestions:   []string{"Q1"},
		SuggestionsEnabled:   true,
	}
	id, err := CreateChatbot(context.Background(), db, b)
	if err != nil || id == "" {
		t.Fatalf("create chatbot: %v", err)
	}
	// update cleanup closures with created id
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM messages WHERE conversation_id IN (SELECT id FROM conversations WHERE chatbot_id = $1)`, id)
		_, _ = db.Exec(`DELETE FROM conversations WHERE chatbot_id = $1`, id)
		_, _ = db.Exec(`DELETE FROM data_sources WHERE chatbot_id = $1`, id)
		_, _ = db.Exec(`DELETE FROM chatbots WHERE id = $1`, id)
	})
	got, err := GetChatbotByID(context.Background(), db, id)
	if err != nil || got == nil {
		t.Fatalf("get chatbot: %v", err)
	}
	if len(got.SuggestedQuestions) != 1 {
		t.Fatalf("suggestions not read")
	}
	// update suggestions
	if err2 := UpdateChatbotSuggestions(context.Background(), db, id, []string{"A", "B"}); err2 != nil {
		t.Fatalf("update sugg: %v", err2)
	}
	got2, err2 := GetChatbotByID(context.Background(), db, id)
	if err2 != nil || got2 == nil {
		t.Fatalf("get2: %v", err2)
	}
	if len(got2.SuggestedQuestions) != 2 {
		t.Fatalf("suggestions not updated")
	}
	// soft delete
	if _, err3 := SoftDeleteChatbot(context.Background(), db, id, uid); err3 != nil {
		t.Fatalf("soft delete: %v", err3)
	}
	got3, err3 := GetChatbotByID(context.Background(), db, id)
	if err3 != nil {
		t.Fatalf("get3: %v", err3)
	}
	if got3 != nil {
		t.Fatalf("expected nil after delete")
	}
}
