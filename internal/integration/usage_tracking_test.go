package integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/policy"
)

// Helper to insert user
func insertUser(t *testing.T, pool *sql.DB, email string) (string, string) {
	var id string
	err := pool.QueryRow(`
        INSERT INTO users (email, password_hash, full_name, plan_id)
        VALUES ($1, 'hash', 'Test User', (SELECT id FROM plans WHERE code=$2))
        RETURNING id`, email, policy.PlanFree.String()).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	return id, email
}

// Helper to insert chatbot
func insertChatbot(t *testing.T, pool *sql.DB, userID string, name string) (string, string) {
	var id string
	err := pool.QueryRow(`
        INSERT INTO chatbots (user_id, name, model)
        VALUES ($1, $2, $3)
        RETURNING id`, userID, name, policy.ModelGPT4oMini.String()).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert chatbot: %v", err)
	}
	return id, name
}

// USG-001 to USG-005: Monthly Token Usage Tracking
func TestMonthlyTokenUsageTracking(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Create user
	userID, _ := insertUser(t, te.DB, "usg-test@example.com")

	// Create two chatbots
	bot1ID, _ := insertChatbot(t, te.DB, userID, "Bot 1")
	bot2ID, _ := insertChatbot(t, te.DB, userID, "Bot 2")

	ctx := context.Background()
	now := time.Now()

	// USG-002: IncrementAnalytics upsert (initial)
	// Add 100 tokens to Bot 1
	err = db.IncrementAnalytics(ctx, te.DB, bot1ID, now, true, 100, false, 500)
	if err != nil {
		t.Fatalf("IncrementAnalytics failed: %v", err)
	}

	// USG-001: GetMonthlyTokenUsage aggregates from analytics
	used, err := db.GetMonthlyTokenUsage(ctx, te.DB, userID)
	if err != nil {
		t.Fatalf("GetMonthlyTokenUsage failed: %v", err)
	}
	if used != 100 {
		t.Errorf("USG-001: expected 100 tokens, got %d", used)
	}

	// USG-003: IncrementAnalytics upsert (update)
	// Add 50 more tokens to Bot 1
	err = db.IncrementAnalytics(ctx, te.DB, bot1ID, now, false, 50, false, 200)
	if err != nil {
		t.Fatalf("IncrementAnalytics update failed: %v", err)
	}

	used, _ = db.GetMonthlyTokenUsage(ctx, te.DB, userID)
	if used != 150 {
		t.Errorf("USG-003: expected 150 tokens, got %d", used)
	}

	// USG-004: Token usage across multiple chatbots
	// Add 200 tokens to Bot 2
	err = db.IncrementAnalytics(ctx, te.DB, bot2ID, now, true, 200, false, 600)
	if err != nil {
		t.Fatalf("IncrementAnalytics bot2 failed: %v", err)
	}

	used, _ = db.GetMonthlyTokenUsage(ctx, te.DB, userID)
	if used != 350 { // 150 + 200
		t.Errorf("USG-004: expected 350 tokens, got %d", used)
	}

	// USG-005: Usage resets at month boundary
	// Add usage for previous month
	// Use explicit date to avoid issues near month boundaries
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prevMonth := currentMonth.AddDate(0, -1, 15) // 15th of previous month

	err = db.IncrementAnalytics(ctx, te.DB, bot1ID, prevMonth, true, 1000, false, 500)
	if err != nil {
		t.Fatalf("IncrementAnalytics prev month failed: %v", err)
	}

	// Should still be 350 for current month
	used, _ = db.GetMonthlyTokenUsage(ctx, te.DB, userID)
	if used != 350 {
		t.Errorf("USG-005: expected 350 tokens (ignoring prev month), got %d", used)
	}
}

// USG-006 to USG-008: Ingestion Usage Tracking
func TestIngestionUsageTracking(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	userID, _ := insertUser(t, te.DB, "ingest-usg@example.com")
	ctx := context.Background()
	now := time.Now()

	// Initial check
	sources, tokens, err := db.GetMonthlyIngestionUsage(ctx, te.DB, userID, now)
	if err != nil {
		t.Fatalf("GetMonthlyIngestionUsage failed: %v", err)
	}
	if sources != 0 || tokens != 0 {
		t.Errorf("expected 0/0, got %d/%d", sources, tokens)
	}

	// USG-007: IncrementSuccessfulIngestion
	err = db.IncrementSuccessfulIngestion(ctx, te.DB, userID, now, 1)
	if err != nil {
		t.Fatalf("IncrementSuccessfulIngestion failed: %v", err)
	}

	var incrementedSources int
	incrementedSources, _, _ = db.GetMonthlyIngestionUsage(ctx, te.DB, userID, now)
	if incrementedSources != 1 {
		t.Errorf("USG-007: expected 1 source, got %d", incrementedSources)
	}

	// USG-008: AddEmbeddingTokens
	err = db.AddEmbeddingTokens(ctx, te.DB, userID, now, 500)
	if err != nil {
		t.Fatalf("AddEmbeddingTokens failed: %v", err)
	}

	_, tokens, _ = db.GetMonthlyIngestionUsage(ctx, te.DB, userID, now)
	if tokens != 500 {
		t.Errorf("USG-008: expected 500 tokens, got %d", tokens)
	}

	// USG-006: GetMonthlyIngestionUsage returns sources + embedding tokens
	// Add more to verify aggregation
	_ = db.IncrementSuccessfulIngestion(ctx, te.DB, userID, now, 2)
	_ = db.AddEmbeddingTokens(ctx, te.DB, userID, now, 1000)

	var embeddingSources int
	embeddingSources, tokens, _ = db.GetMonthlyIngestionUsage(ctx, te.DB, userID, now)
	if embeddingSources != 3 || tokens != 1500 {
		t.Errorf("USG-006: expected 3/1500, got %d/%d", embeddingSources, tokens)
	}
}
