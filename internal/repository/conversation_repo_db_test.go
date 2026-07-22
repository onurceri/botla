package repository_test

import (
	"context"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/testdb"
)

// TestSaveMessageSources_BulkInsert verifies that SaveMessageSources persists all
// sources in one statement (including the source_type column added in migration
// 000004), skips entries with an empty SourceID, and de-duplicates via
// ON CONFLICT DO NOTHING on a second save.
func TestSaveMessageSources_BulkInsert(t *testing.T) {
	db := testdb.OpenTestDB(t)
	repo := repository.NewPostgresConversationRepo(db)
	ctx := context.Background()

	// CreateSource builds the full hierarchy (user/org/workspace/chatbot) + a data_source.
	srcResult := testdb.CreateSource(t, db)
	chatbotID := srcResult.Chatbot.ID
	sourceID := srcResult.Source.ID

	conv, err := repo.GetOrCreateBySessionID(ctx, chatbotID, "sess-"+sourceID)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	msgID, err := repo.CreateMessage(ctx, &models.Message{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        "answer",
	})
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	sources := []models.ChunkMetadata{
		{SourceID: sourceID, SourceType: "text", ChunkIndex: 0, Score: 0.95},
		{SourceID: sourceID, SourceType: "text", ChunkIndex: 1, Score: 0.80},
		{SourceID: "", ChunkIndex: 2, Score: 0.5}, // skipped: empty SourceID
	}
	if err := repo.SaveMessageSources(ctx, msgID, sources); err != nil {
		t.Fatalf("SaveMessageSources: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM message_sources WHERE message_id = $1", msgID).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rows (empty SourceID skipped), got %d", count)
	}

	var gotType string
	var gotScore float64
	if err := db.QueryRowContext(ctx,
		"SELECT source_type, relevance_score FROM message_sources WHERE message_id = $1 AND chunk_index = 0", msgID,
	).Scan(&gotType, &gotScore); err != nil {
		t.Fatalf("read row: %v", err)
	}
	if gotType != "text" {
		t.Fatalf("source_type = %q, want %q", gotType, "text")
	}
	if gotScore != 0.95 {
		t.Fatalf("relevance_score = %v, want 0.95", gotScore)
	}

	// ON CONFLICT DO NOTHING: re-saving the same sources must not duplicate rows.
	if err := repo.SaveMessageSources(ctx, msgID, sources); err != nil {
		t.Fatalf("re-save: %v", err)
	}
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM message_sources WHERE message_id = $1", msgID).Scan(&count); err != nil {
		t.Fatalf("recount: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected still 2 rows after re-save (ON CONFLICT DO NOTHING), got %d", count)
	}
}
