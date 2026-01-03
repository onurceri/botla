package integration

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

// R2-006: Storage used MB tracking
func TestStorageUsageTracking(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	userID, _ := insertUser(t, te.DB, "storage-usg@example.com")
	botID, _ := insertChatbot(t, te.DB, userID, "Storage Bot")

	ctx := context.Background()

	// Initial check: 0 MB
	used, err := db.GetStorageUsedMBByUserID(ctx, te.DB, userID)
	if err != nil {
		t.Fatalf("GetStorageUsedMBByUserID failed: %v", err)
	}
	if used != 0 {
		t.Errorf("expected 0 MB, got %d", used)
	}

	// Insert a source with 5 MB
	size1 := 5 * 1024 * 1024
	_, err = te.DB.Exec(`INSERT INTO data_sources (chatbot_id, original_filename, source_type, status, size_bytes) VALUES ($1, 's1', 'text', 'processed', $2)`, botID, size1)
	if err != nil {
		t.Fatalf("insert source 1 failed: %v", err)
	}

	// Check
	used, _ = db.GetStorageUsedMBByUserID(ctx, te.DB, userID)
	if used != 5 {
		t.Errorf("expected 5 MB, got %d", used)
	}

	// Insert another source with 3.5 MB
	// Total: 8.5 MB -> 8 MB (integer division)
	size2 := int(3.5 * 1024 * 1024)
	_, err = te.DB.Exec(`INSERT INTO data_sources (chatbot_id, original_filename, source_type, status, size_bytes) VALUES ($1, 's2', 'text', 'processed', $2)`, botID, size2)
	if err != nil {
		t.Fatalf("insert source 2 failed: %v", err)
	}

	used, _ = db.GetStorageUsedMBByUserID(ctx, te.DB, userID)
	if used != 8 {
		t.Errorf("expected 8 MB, got %d", used)
	}

	// Deleted source should not count
	_, err = te.DB.Exec(`UPDATE data_sources SET deleted_at = NOW() WHERE original_filename = 's1'`)
	if err != nil {
		t.Fatalf("delete source 1 failed: %v", err)
	}

	// Should be 3 MB (3.5 rounded down)
	used, _ = db.GetStorageUsedMBByUserID(ctx, te.DB, userID)
	if used != 3 {
		t.Errorf("expected 3 MB, got %d", used)
	}
}
