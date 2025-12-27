package db_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
)

func TestSourceExistsByHash(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)

	cbResult := testdb.CreateChatbot(t, dbConn)
	chatbotID := cbResult.Chatbot.ID

	// Create a source with hash
	hash := "abc123hash"
	createSourceWithHash(t, dbConn, chatbotID, hash)

	// Test: Same chatbot, same hash should exist
	exists, err := db.SourceExistsByHash(context.Background(), dbConn, chatbotID, hash)
	if err != nil {
		t.Fatalf("SourceExistsByHash failed: %v", err)
	}
	if !exists {
		t.Error("expected source to exist")
	}

	// Test: Same chatbot, different hash should not exist
	exists, err = db.SourceExistsByHash(context.Background(), dbConn, chatbotID, "different-hash")
	if err != nil {
		t.Fatalf("SourceExistsByHash failed: %v", err)
	}
	if exists {
		t.Error("expected source to not exist")
	}

	// Test: Different chatbot, same hash should not exist (ok to have same content in different bots)
	cbResult2 := testdb.CreateChatbot(t, dbConn)
	otherChatbotID := cbResult2.Chatbot.ID
	exists, err = db.SourceExistsByHash(context.Background(), dbConn, otherChatbotID, hash)
	if err != nil {
		t.Fatalf("SourceExistsByHash failed: %v", err)
	}
	if exists {
		t.Error("expected source to not exist in different chatbot")
	}
}

func TestSourceExistsByHash_DeletedSource(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)

	cbResult := testdb.CreateChatbot(t, dbConn)
	chatbotID := cbResult.Chatbot.ID
	hash := "deleted-hash"

	// Create and delete a source
	sourceID := createSourceWithHash(t, dbConn, chatbotID, hash)
	err := db.SoftDeleteSource(context.Background(), dbConn, sourceID)
	if err != nil {
		t.Fatalf("SoftDeleteSource failed: %v", err)
	}

	// Should NOT find deleted source
	exists, err := db.SourceExistsByHash(context.Background(), dbConn, chatbotID, hash)
	if err != nil {
		t.Fatalf("SourceExistsByHash failed: %v", err)
	}
	if exists {
		t.Error("should not find deleted source")
	}
}

func TestGetSourceByHash(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)

	cbResult := testdb.CreateChatbot(t, dbConn)
	chatbotID := cbResult.Chatbot.ID
	hash := "get-by-hash"

	sourceID := createSourceWithHash(t, dbConn, chatbotID, hash)

	// Test: Found
	source, err := db.GetSourceByHash(context.Background(), dbConn, chatbotID, hash)
	if err != nil {
		t.Fatalf("GetSourceByHash failed: %v", err)
	}
	if source == nil {
		t.Fatal("expected source to be found")
	}
	if source.ID != sourceID {
		t.Errorf("expected source ID %s, got %s", sourceID, source.ID)
	}

	// Test: Not found (different hash)
	source, err = db.GetSourceByHash(context.Background(), dbConn, chatbotID, "not-exist")
	if err != nil {
		t.Fatalf("GetSourceByHash failed: %v", err)
	}
	if source != nil {
		t.Error("expected source to be nil")
	}

	// Test: Not found (deleted)
	err = db.SoftDeleteSource(context.Background(), dbConn, sourceID)
	if err != nil {
		t.Fatalf("SoftDeleteSource failed: %v", err)
	}
	source, err = db.GetSourceByHash(context.Background(), dbConn, chatbotID, hash)
	if err != nil {
		t.Fatalf("GetSourceByHash failed: %v", err)
	}
	if source != nil {
		t.Error("expected deleted source to not be found")
	}
}

func createSourceWithHash(t *testing.T, dbConn *sql.DB, chatbotID, hash string) string {
	t.Helper()
	ds := &models.DataSource{
		ChatbotID:  chatbotID,
		SourceType: "text",
		Status:     "completed",
		Hash:       &hash,
	}
	id, err := db.CreateDataSource(context.Background(), dbConn, ds)
	if err != nil {
		t.Fatalf("failed to create source with hash: %v", err)
	}
	return id
}
