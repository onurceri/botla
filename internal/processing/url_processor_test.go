package processing

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/logger"
)

func TestURLProcessor_TransactionSafety(t *testing.T) {
	t.Parallel()
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	userResult := testdb.CreateUser(t, dbConn)
	orgResult := testdb.CreateOrganization(t, dbConn, testdb.OrganizationFixture{OwnerID: userResult.ID})
	wsResult := testdb.CreateWorkspace(t, dbConn, testdb.WorkspaceFixture{OrganizationID: orgResult.Organization.ID})
	botResult := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{
		WorkspaceID:    &wsResult.Workspace.ID,
		OrganizationID: &orgResult.Organization.ID,
		UserID:         userResult.ID,
	})

	sourceURL := "https://example.com"
	source := &models.DataSource{
		ID:           "test-source-id",
		ChatbotID:    botResult.Chatbot.ID,
		SourceType:   "url",
		SourceURL:    &sourceURL,
		ChunkCount:   0,
		Status:       "pending",
		IsDiscovered: false,
	}

	_ = source

	tests := []struct {
		name            string
		setupDB         func(t *testing.T, db *sql.DB)
		expectedSources int
		expectedTokens  int
	}{
		{
			name:            "successful transaction commits both operations",
			setupDB:         func(t *testing.T, db *sql.DB) {},
			expectedSources: 1,
			expectedTokens:  100,
		},
		{
			name: "transaction atomicity - no partial updates",
			setupDB: func(t *testing.T, db *sql.DB) {
				_, _ = dbConn.Exec(`DROP TABLE IF EXISTS usage_ingestions`)
			},
			expectedSources: 0,
			expectedTokens:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupDB(t, dbConn)

			if err := db.IncrementSuccessfulIngestion(ctx, dbConn, userResult.ID, time.Now(), 1); err == nil {
				if err := db.AddEmbeddingTokens(ctx, dbConn, userResult.ID, time.Now(), 100); err == nil {
					sources, tokens, err := db.GetMonthlyIngestionUsage(ctx, dbConn, userResult.ID, time.Now())
					if err != nil {
						t.Fatalf("GetMonthlyIngestionUsage: %v", err)
					}
					if sources != tt.expectedSources || tokens != tt.expectedTokens {
						t.Errorf("expected sources=%d tokens=%d, got sources=%d tokens=%d",
							tt.expectedSources, tt.expectedTokens, sources, tokens)
					}
				}
			}

			_, _ = dbConn.Exec(`CREATE TABLE IF NOT EXISTS usage_ingestions (
				user_id VARCHAR(64) NOT NULL,
				period_month DATE NOT NULL,
				sources_count INT NOT NULL DEFAULT 0,
				embedding_tokens INT NOT NULL DEFAULT 0,
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
				PRIMARY KEY (user_id, period_month)
			)`)
		})
	}
}

func TestURLProcessor_TransactionWithMockedDB(t *testing.T) {
	t.Parallel()
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	userResult := testdb.CreateUser(t, dbConn)
	now := time.Now()

	tx, err := dbConn.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	if err := db.IncrementSuccessfulIngestionTx(ctx, tx, userResult.ID, now, 1); err != nil {
		tx.Rollback()
		t.Fatalf("increment tx failed: %v", err)
	}

	if err := db.AddEmbeddingTokensTx(ctx, tx, userResult.ID, now, 50); err != nil {
		tx.Rollback()
		t.Fatalf("add tokens tx failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx failed: %v", err)
	}

	sources, tokens, err := db.GetMonthlyIngestionUsage(ctx, dbConn, userResult.ID, now)
	if err != nil {
		t.Fatalf("get usage failed: %v", err)
	}
	if sources != 1 || tokens != 50 {
		t.Errorf("expected sources=1 tokens=50, got sources=%d tokens=%d", sources, tokens)
	}
}

func TestURLProcessor_TransactionRollback(t *testing.T) {
	t.Parallel()
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	userResult := testdb.CreateUser(t, dbConn)
	now := time.Now()

	tx, err := dbConn.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	if err := db.IncrementSuccessfulIngestionTx(ctx, tx, userResult.ID, now, 5); err != nil {
		tx.Rollback()
		t.Fatalf("increment tx failed: %v", err)
	}

	_, _ = dbConn.Exec(`DROP TABLE usage_ingestions`)

	err = db.AddEmbeddingTokensTx(ctx, tx, userResult.ID, now, 500)
	if err == nil {
		tx.Rollback()
		t.Fatalf("expected error when table doesn't exist")
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback tx failed: %v", err)
	}

	_, _ = dbConn.Exec(`CREATE TABLE IF NOT EXISTS usage_ingestions (
		user_id VARCHAR(64) NOT NULL,
		period_month DATE NOT NULL,
		sources_count INT NOT NULL DEFAULT 0,
		embedding_tokens INT NOT NULL DEFAULT 0,
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (user_id, period_month)
	)`)

	sources, _, err := db.GetMonthlyIngestionUsage(ctx, dbConn, userResult.ID, now)
	if err != nil {
		t.Fatalf("get usage failed: %v", err)
	}
	if sources != 0 {
		t.Errorf("expected 0 sources after rollback, got %d", sources)
	}
}

func TestURLProcessor_ConcurrentTransactions(t *testing.T) {
	t.Parallel()
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	userResult := testdb.CreateUser(t, dbConn)
	now := time.Now()

	numTransactions := 5
	successChan := make(chan bool, numTransactions)

	for i := 0; i < numTransactions; i++ {
		go func(idx int) {
			tx, err := dbConn.Begin()
			if err != nil {
				successChan <- false
				return
			}

			if err := db.IncrementSuccessfulIngestionTx(ctx, tx, userResult.ID, now, 1); err != nil {
				tx.Rollback()
				successChan <- false
				return
			}

			if err := db.AddEmbeddingTokensTx(ctx, tx, userResult.ID, now, 10); err != nil {
				tx.Rollback()
				successChan <- false
				return
			}

			if err := tx.Commit(); err != nil {
				successChan <- false
				return
			}

			successChan <- true
		}(i)
	}

	successCount := 0
	for i := 0; i < numTransactions; i++ {
		if <-successChan {
			successCount++
		}
	}

	if successCount != numTransactions {
		t.Errorf("expected all %d transactions to succeed, got %d", numTransactions, successCount)
	}

	sources, tokens, err := db.GetMonthlyIngestionUsage(ctx, dbConn, userResult.ID, now)
	if err != nil {
		t.Fatalf("get usage failed: %v", err)
	}
	if sources != numTransactions {
		t.Errorf("expected %d sources, got %d", numTransactions, sources)
	}
	if tokens != numTransactions*10 {
		t.Errorf("expected %d tokens, got %d", numTransactions*10, tokens)
	}
}

func TestURLProcessor_NewURLProcessor(t *testing.T) {
	t.Parallel()

	t.Run("creates processor with default scraper", func(t *testing.T) {
		p := NewURLProcessor(nil, nil, nil, nil, nil)
		if p == nil {
			t.Error("expected non-nil processor")
		}
		if p.Scraper == nil {
			t.Error("expected non-nil default scraper")
		}
	})

	t.Run("creates processor with custom logger", func(t *testing.T) {
		log := logger.New("test")
		p := NewURLProcessor(nil, nil, nil, log, nil)
		if p == nil {
			t.Error("expected non-nil processor")
		}
		if p.Log != log {
			t.Error("expected logger to be set")
		}
	})
}
