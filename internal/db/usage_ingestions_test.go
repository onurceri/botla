package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/testdb"
)

func TestUsageIngestions_CRUD(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	_, _ = dbConn.Exec(`CREATE TABLE IF NOT EXISTS usage_ingestions (
        user_id VARCHAR(64) NOT NULL,
        period_month DATE NOT NULL,
        sources_count INT NOT NULL DEFAULT 0,
        embedding_tokens INT NOT NULL DEFAULT 0,
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        PRIMARY KEY (user_id, period_month)
    )`)
	uid := createUser(t, dbConn)
	now := time.Now()
	if err := db.IncrementSuccessfulIngestion(context.Background(), dbConn, uid, now, 2); err != nil {
		t.Fatalf("inc: %v", err)
	}
	if err := db.AddEmbeddingTokens(context.Background(), dbConn, uid, now, 123); err != nil {
		t.Fatalf("tokens: %v", err)
	}
	s, tok, err := db.GetMonthlyIngestionUsage(context.Background(), dbConn, uid, now)
	if err != nil || s != 2 || tok != 123 {
		t.Fatalf("get: %v s=%d tok=%d", err, s, tok)
	}
}
