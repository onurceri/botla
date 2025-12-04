package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/internal/models"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

func createUser(t *testing.T, db *sql.DB) string {
	t.Helper()
	email := fmt.Sprintf("dbu+%d@example.com", time.Now().UnixNano())
	var id string
	if err := db.QueryRow(`INSERT INTO users (email, password_hash) VALUES ($1,$2) RETURNING id`, email, "x").Scan(&id); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}

func TestChatbot_CRUD_DB(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	uid := createUser(t, db)
	b := &models.Chatbot{
		UserID:               uid,
		Name:                 "DB Bot",
		SystemPrompt:         "p",
		Language:             "en",
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
	got, err := GetChatbotByID(context.Background(), db, id)
	if err != nil || got == nil {
		t.Fatalf("get chatbot: %v", err)
	}
	if len(got.SuggestedQuestions) != 1 {
		t.Fatalf("suggestions not read")
	}
	// update suggestions
	if err := UpdateChatbotSuggestions(context.Background(), db, id, []string{"A", "B"}); err != nil {
		t.Fatalf("update sugg: %v", err)
	}
	got2, err := GetChatbotByID(context.Background(), db, id)
	if err != nil || got2 == nil {
		t.Fatalf("get2: %v", err)
	}
	if len(got2.SuggestedQuestions) != 2 {
		t.Fatalf("suggestions not updated")
	}
	// soft delete
	if err := SoftDeleteChatbot(context.Background(), db, id, uid); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	got3, err := GetChatbotByID(context.Background(), db, id)
	if err != nil {
		t.Fatalf("get3: %v", err)
	}
	if got3 != nil {
		t.Fatalf("expected nil after delete")
	}
}
