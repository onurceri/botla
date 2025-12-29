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

func TestUsageIngestions_Tx_Success(t *testing.T) {
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

	tx, err := dbConn.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	if err := db.IncrementSuccessfulIngestionTx(context.Background(), tx, uid, now, 5); err != nil {
		tx.Rollback()
		t.Fatalf("increment tx: %v", err)
	}
	if err := db.AddEmbeddingTokensTx(context.Background(), tx, uid, now, 500); err != nil {
		tx.Rollback()
		t.Fatalf("add tokens tx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	s, tok, err := db.GetMonthlyIngestionUsage(context.Background(), dbConn, uid, now)
	if err != nil || s != 5 || tok != 500 {
		t.Fatalf("get after tx: %v s=%d tok=%d", err, s, tok)
	}
}

func TestUsageIngestions_Tx_RollbackOnFailure(t *testing.T) {
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

	tx, err := dbConn.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	if err := db.IncrementSuccessfulIngestionTx(context.Background(), tx, uid, now, 10); err != nil {
		tx.Rollback()
		t.Fatalf("increment tx: %v", err)
	}

	_, _ = dbConn.Exec(`DROP TABLE usage_ingestions`)

	err = db.AddEmbeddingTokensTx(context.Background(), tx, uid, now, 1000)
	if err == nil {
		tx.Rollback()
		t.Fatalf("expected error when table doesn't exist, but got none")
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback tx: %v", err)
	}

	_, _ = dbConn.Exec(`CREATE TABLE IF NOT EXISTS usage_ingestions (
        user_id VARCHAR(64) NOT NULL,
        period_month DATE NOT NULL,
        sources_count INT NOT NULL DEFAULT 0,
        embedding_tokens INT NOT NULL DEFAULT 0,
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        PRIMARY KEY (user_id, period_month)
    )`)

	s, _, err := db.GetMonthlyIngestionUsage(context.Background(), dbConn, uid, now)
	if err != nil {
		t.Fatalf("get usage: %v", err)
	}
	if s != 0 {
		t.Fatalf("expected 0 sources after rollback, got %d", s)
	}
}

func TestUsageIngestions_Tx_Atomicity(t *testing.T) {
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

	tx, err := dbConn.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	if err := db.IncrementSuccessfulIngestionTx(context.Background(), tx, uid, now, 3); err != nil {
		tx.Rollback()
		t.Fatalf("increment tx: %v", err)
	}

	_, _ = dbConn.Exec(`DROP TABLE usage_ingestions`)

	err = db.AddEmbeddingTokensTx(context.Background(), tx, uid, now, 300)
	if err == nil {
		tx.Rollback()
		t.Fatalf("expected error when table doesn't exist, but got none")
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback tx: %v", err)
	}

	_, _ = dbConn.Exec(`CREATE TABLE IF NOT EXISTS usage_ingestions (
        user_id VARCHAR(64) NOT NULL,
        period_month DATE NOT NULL,
        sources_count INT NOT NULL DEFAULT 0,
        embedding_tokens INT NOT NULL DEFAULT 0,
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        PRIMARY KEY (user_id, period_month)
    )`)

	s, tok, err := db.GetMonthlyIngestionUsage(context.Background(), dbConn, uid, now)
	if err != nil {
		t.Fatalf("get usage: %v", err)
	}
	if s != 0 || tok != 0 {
		t.Fatalf("expected 0 sources and 0 tokens after rollback, got s=%d tok=%d", s, tok)
	}
}

func TestUsageIngestions_Tx_Consecutive(t *testing.T) {
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

	for i := 0; i < 3; i++ {
		tx, err := dbConn.Begin()
		if err != nil {
			t.Fatalf("begin tx %d: %v", i, err)
		}

		if err := db.IncrementSuccessfulIngestionTx(context.Background(), tx, uid, now, 1); err != nil {
			tx.Rollback()
			t.Fatalf("increment tx %d: %v", i, err)
		}
		if err := db.AddEmbeddingTokensTx(context.Background(), tx, uid, now, 100); err != nil {
			tx.Rollback()
			t.Fatalf("add tokens tx %d: %v", i, err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit tx %d: %v", i, err)
		}
	}

	s, tok, err := db.GetMonthlyIngestionUsage(context.Background(), dbConn, uid, now)
	if err != nil {
		t.Fatalf("get after consecutive txs: %v", err)
	}
	if s != 3 || tok != 300 {
		t.Fatalf("expected 3 sources and 300 tokens, got s=%d tok=%d", s, tok)
	}
}
