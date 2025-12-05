package db

import (
	"context"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/internal/models"
)

func TestDataSource_CRUD_DB(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	uid := createUser(t, db)
	b := &models.Chatbot{UserID: uid, Name: "Src Bot", SystemPrompt: "p", LanguageCode: "en-US", Model: "gpt-3.5-turbo", Temperature: 0.1, MaxTokens: 64, ThemeColor: "#000000", WelcomeMessage: "hi", Position: "bottom-right", BotMessageColor: "#000000", UserMessageColor: "#ffffff", BotMessageTextColor: "#ffffff", UserMessageTextColor: "#000000", ChatFontFamily: "Inter", ChatHeaderColor: "#000000", ChatHeaderTextColor: "#ffffff", ChatBackgroundColor: "#ffffff"}
	bid, err := CreateChatbot(context.Background(), db, b)
	if err != nil {
		t.Fatalf("create bot: %v", err)
	}
	// create source
	s := &models.DataSource{ChatbotID: bid, SourceType: "text", Status: "pending", ChunkCount: 0}
	sid, err := CreateDataSource(context.Background(), db, s)
	if err != nil || sid == "" {
		t.Fatalf("create source: %v", err)
	}
	// list
	list, err := ListSourcesByChatbotID(context.Background(), db, bid)
	if err != nil || len(list) == 0 {
		t.Fatalf("list sources: %v", err)
	}
	// update processing
	now := time.Now()
	em := ""
	if err := UpdateSourceProcessing(context.Background(), db, sid, "completed", &em, 3, &now); err != nil {
		t.Fatalf("update processing: %v", err)
	}
	got, err := GetSourceByID(context.Background(), db, sid)
	if err != nil || got == nil || got.Status != "completed" {
		t.Fatalf("get source: %v", err)
	}
	// update capability and suggestions
	if err := UpdateSourceCapability(context.Background(), db, sid, "summary"); err != nil {
		t.Fatalf("cap: %v", err)
	}
	if err := UpdateSourceSuggestions(context.Background(), db, sid, []string{"q1", "q2"}); err != nil {
		t.Fatalf("sugg: %v", err)
	}
	// delete
	if err := DeleteSource(context.Background(), db, sid); err != nil {
		t.Fatalf("delete: %v", err)
	}
	gone, err := GetSourceByID(context.Background(), db, sid)
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if gone != nil {
		t.Fatalf("expected nil after delete")
	}
}
