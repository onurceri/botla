package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/policy"
)

func TestChatService_ProcessChatWithValidation_TokenQuotaExceeded(t *testing.T) {
	t.Parallel()
	dbConn := testdb.OpenParallelTestDB(t)

	// Create a plan with strict token limit
	// Note: We need to update the user's plan to something with limits.
	// Since creating a new plan type in DB test setup is hard (migration dependent),
	// we will manually update the 'free' plan or similar in the test transaction/schema
	// OR insert a custom plan if constraints allow.

	// Better: Update the plan config for the user's plan.
	// The test schema migrations seed plans. Let's assume 'free' exists.
	// We'll update 'free' plan to have specific MaxMonthlyTokens.

	_, err := dbConn.Exec(`
		UPDATE plans 
		SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '100') 
		WHERE code = $1
	`, policy.PlanFree.String())
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	// Create User on 'free' plan
	user := testdb.CreateUser(t, dbConn, testdb.UserFixture{
		PlanCode: policy.PlanFree.String(),
	})

	// Create Chatbot
	cbResult := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{
		UserID:    user.ID,
		MaxTokens: 50, // Requesting 50 tokens
	})

	// Manually insert usage to reach quota
	// Quota is 100. Let's insert 100 usage.
	// We need to insert into usage_ingestions or usage_chat (wait, where is chat usage tracked?)
	// The user memory "Fix Token Quota Race Condition" said "adding a chat_tokens column to the usage_ingestions table"
	// OR "tracking monthly chat token usage per user".
	// Actually logic uses `db.GetMonthlyTokenUsage` and `db.ReserveChatTokens`.
	// `ReserveChatTokens` checks `usage_chat` or similar.
	// Let's assume ReserveChatTokens works if usage is high.
	// We can cheat by calling ReserveChatTokens manually to use up quota.

	err = db.ReserveChatTokens(context.Background(), dbConn, user.ID, 100, 100)
	if err != nil {
		t.Fatalf("failed to setup initial quota usage: %v", err)
	}

	// Now we have used 100/100 tokens.
	// Try to process chat requesting 50 tokens (Chatbot.MaxTokens).

	svc := services.NewChatService(dbConn, rag.NewClientFactory(&config.Config{}), nil, nil, logger.New("ERROR"))

	req := services.ChatRequestWithUser{
		UserID:      user.ID,
		Chatbot:     cbResult.Chatbot,
		ChatRequest: models.ChatRequest{Message: "Hello", SessionID: "sess-1"},
	}

	_, err = svc.ProcessChatWithValidation(context.Background(), req)

	if !errors.Is(err, services.ErrTokenQuotaExceeded) {
		t.Errorf("ProcessChatWithValidation() error = %v, want ErrTokenQuotaExceeded", err)
	}
}

func TestChatService_ProcessChatWithValidation_DelegationAndRefund(t *testing.T) {
	t.Parallel()
	dbConn := testdb.OpenParallelTestDB(t)

	// Update plan to have quota
	_, err := dbConn.Exec(`
		UPDATE plans 
		SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '1000') 
		WHERE code = $1
	`, policy.PlanFree.String())
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	user := testdb.CreateUser(t, dbConn, testdb.UserFixture{PlanCode: policy.PlanFree.String()})
	cbResult := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{
		UserID:    user.ID,
		MaxTokens: 100,
	})

	svc := services.NewChatService(dbConn, rag.NewClientFactory(&config.Config{}), nil, nil, logger.New("ERROR"))

	req := services.ChatRequestWithUser{
		UserID:      user.ID,
		Chatbot:     cbResult.Chatbot,
		ChatRequest: models.ChatRequest{Message: "Hello", SessionID: "sess-1"},
	}

	// This should fail inside ProcessChat (due to no keys/models),
	// but it should PASS validation (quota 100 requested < 1000 limit).
	// Because ProcessChat fails, it should refund the tokens.

	_, err = svc.ProcessChatWithValidation(context.Background(), req)

	// Error should NOT be TokenQuotaExceeded
	if errors.Is(err, services.ErrTokenQuotaExceeded) {
		t.Fatalf("got TokenQuotaExceeded unexpectedly")
	}
	if err == nil {
		// If by miracle it succeeds (no keys?), that's weird but ok if mocks were injected.
		// But valid test env expects failure.
		t.Log("ProcessChat unexpectedly succeeded")
	} else {
		t.Logf("ProcessChat failed as expected with: %v", err)
	}

	// Verify refund: Usage should be 0 (or whatever it was before).
	// usage_ingestions table tracks chat_tokens.
	// We can check by calling ReserveChatTokens again for 1000. It should succeed if refund worked.
	// Or check usage directly. db.GetMonthlyTokenUsage is internal?
	// `db.GetMonthlyTokenUsage` is exported from `db` package? No, it's likely generated.
	// We can use SQL to check.
	// The table `usage_ingestions` likely has `chat_tokens` column (sum of it).
	// Wait, `ReserveChatTokens` logic is complex.
	// It likely inserts a row. Refund updates it?
	// `AdjustChatTokens` updates the row.

	var totalTokens int
	err = dbConn.QueryRow(`
		SELECT COALESCE(SUM(chat_tokens), 0) 
		FROM usage_ingestions 
		WHERE user_id = $1
	`, user.ID).Scan(&totalTokens)
	if err != nil {
		t.Fatalf("failed to query usage: %v", err)
	}

	if totalTokens != 0 {
		t.Errorf("expected 0 tokens used after refund, got %d", totalTokens)
	}
}
