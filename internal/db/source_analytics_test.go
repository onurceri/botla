package db

import (
	"context"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
)

func TestGetSourceUsageStats(t *testing.T) {
	db := testdb.OpenTestDB(t)
	defer db.Close()

	// Ensure message_sources table exists (in case migration didn't run on test schema)
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS message_sources (
	    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	    message_id UUID NOT NULL,
	    source_id UUID NOT NULL,
	    chunk_index INT NOT NULL,
	    relevance_score FLOAT NOT NULL,
	    created_at TIMESTAMP DEFAULT NOW(),
	    UNIQUE(message_id, source_id, chunk_index)
	)`)

	uid := createUser(t, db)
	// Create Chatbot
	b := &models.Chatbot{UserID: uid, Name: "Analytics Bot", SystemPrompt: "p", LanguageCode: "en-US", Model: "gpt-3.5-turbo", Temperature: 0.1, MaxTokens: 64, ThemeColor: "#000000", WelcomeMessage: "hi", Position: "bottom-right", BotMessageColor: "#000000", UserMessageColor: "#ffffff", BotMessageTextColor: "#ffffff", UserMessageTextColor: "#000000", ChatFontFamily: "Inter", ChatHeaderColor: "#000000", ChatHeaderTextColor: "#ffffff", ChatBackgroundColor: "#ffffff"}
	botID, err := CreateChatbot(context.Background(), db, b)
	if err != nil {
		t.Fatalf("create bot: %v", err)
	}

	// Create Source
	fn := "test.txt"
	s := &models.DataSource{ChatbotID: botID, SourceType: "text", Status: "completed", ChunkCount: 1, OriginalFilename: &fn}
	sourceID, err := CreateDataSource(context.Background(), db, s)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}

	// Create Conversation
	var convID string
	err = db.QueryRow(`INSERT INTO conversations (chatbot_id, session_id) VALUES ($1, 'sess1') RETURNING id`, botID).Scan(&convID)
	if err != nil {
		t.Fatalf("create conv: %v", err)
	}

	// Create Message
	var msgID string
	err = db.QueryRow(`INSERT INTO messages (conversation_id, role, content, thumbs_up) VALUES ($1, 'assistant', 'resp', true) RETURNING id`, convID).Scan(&msgID)
	if err != nil {
		t.Fatalf("create msg: %v", err)
	}

	// Create Message Source (Usage)
	_, err = db.Exec(`INSERT INTO message_sources (message_id, source_id, chunk_index, relevance_score) VALUES ($1, $2, 0, 0.85)`, msgID, sourceID)
	if err != nil {
		t.Fatalf("create msg source: %v", err)
	}

	// Test GetSourceUsageStats
	stats, err := GetSourceUsageStats(context.Background(), db, botID, 30)
	if err != nil {
		t.Fatalf("GetSourceUsageStats failed: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}

	stat := stats[0]
	if stat.SourceID != sourceID {
		t.Errorf("expected sourceID %s, got %s", sourceID, stat.SourceID)
	}
	if stat.TimesUsed != 1 {
		t.Errorf("expected TimesUsed 1, got %d", stat.TimesUsed)
	}
	if stat.AvgRelevance != 0.85 {
		t.Errorf("expected AvgRelevance 0.85, got %f", stat.AvgRelevance)
	}
	if stat.PositiveFeedback != 1 {
		t.Errorf("expected PositiveFeedback 1, got %d", stat.PositiveFeedback)
	}

	if stat.LastUsed.IsZero() {
		t.Error("expected LastUsed to be non-zero")
	}
	if time.Since(stat.LastUsed) > time.Hour {
		t.Error("expected LastUsed to be recent")
	}

	// Clean up
	db.Exec(`DELETE FROM message_sources WHERE message_id = $1`, msgID)
	db.Exec(`DELETE FROM messages WHERE id = $1`, msgID)
	db.Exec(`DELETE FROM conversations WHERE id = $1`, convID)
	db.Exec(`DELETE FROM data_sources WHERE id = $1`, sourceID)
	db.Exec(`DELETE FROM chatbots WHERE id = $1`, botID)
	db.Exec(`DELETE FROM users WHERE id = $1`, uid)
}
