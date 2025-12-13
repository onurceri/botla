package db

import (
	"context"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/testdb"
)

func TestUsageIngestions_CRUD(t *testing.T) {
	db := testdb.OpenTestDB(t)
	defer db.Close()
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS usage_ingestions (
        user_id VARCHAR(64) NOT NULL,
        period_month DATE NOT NULL,
        sources_count INT NOT NULL DEFAULT 0,
        embedding_tokens INT NOT NULL DEFAULT 0,
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        PRIMARY KEY (user_id, period_month)
    )`)
	uid := createUser(t, db)
	now := time.Now()
	if err := IncrementSuccessfulIngestion(context.Background(), db, uid, now, 2); err != nil {
		t.Fatalf("inc: %v", err)
	}
	if err := AddEmbeddingTokens(context.Background(), db, uid, now, 123); err != nil {
		t.Fatalf("tokens: %v", err)
	}
	s, tok, err := GetMonthlyIngestionUsage(context.Background(), db, uid, now)
	if err != nil || s != 2 || tok != 123 {
		t.Fatalf("get: %v s=%d tok=%d", err, s, tok)
	}
}
