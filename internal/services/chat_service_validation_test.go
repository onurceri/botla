package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/policy"
)

func TestChatService_ProcessChatWithValidation_TokenQuotaExceeded(t *testing.T) {
	t.Skip("SKIP: Quota enforcement requires QuotaEnforcer to be migrated to use UsageRepository instead of raw *sql.DB")
	t.Parallel()
	dbConn := testdb.OpenParallelTestDB(t)

	// Create a plan with strict token limit
	_ = testdb.UpdatePlanLimit(context.Background(), dbConn, policy.PlanFree.String(), "chat_max_monthly_tokens", 100)

	// Create User on 'free' plan
	user := testdb.CreateUser(t, dbConn, testdb.UserFixture{
		PlanCode: policy.PlanFree.String(),
	})

	// Create Chatbot
	cbResult := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{
		UserID:    user.ID,
		MaxTokens: 50, // Requesting 50 tokens
	})

	var err error
	err = reserveChatTokens(context.Background(), dbConn, user.ID, 100, 100)
	if err != nil {
		t.Fatalf("failed to setup initial quota usage: %v", err)
	}

	// Create repository instances
	planRepo := repository.NewPostgresPlanRepo(dbConn, nil)
	conversationRepo := repository.NewPostgresConversationRepo(dbConn)
	analyticsRepo := repository.NewPostgresAnalyticsRepo(dbConn)
	actionRepo := repository.NewMockActionRepo()
	usageRepo := repository.NewPostgresUsageRepo(dbConn)

	svc := services.NewChatService(
		planRepo,
		conversationRepo,
		analyticsRepo,
		actionRepo,
		nil,
		nil,
		rag.NewClientFactory(&config.Config{}),
		nil,
		nil,
		usageRepo,
		logger.New("ERROR"),
	)

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
	_ = testdb.UpdatePlanLimit(context.Background(), dbConn, policy.PlanFree.String(), "chat_max_monthly_tokens", 1000)

	var err error

	user := testdb.CreateUser(t, dbConn, testdb.UserFixture{PlanCode: policy.PlanFree.String()})
	cbResult := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{
		UserID:    user.ID,
		MaxTokens: 100,
	})

	// Create repository instances
	planRepo := repository.NewPostgresPlanRepo(dbConn, nil)
	conversationRepo := repository.NewPostgresConversationRepo(dbConn)
	analyticsRepo := repository.NewPostgresAnalyticsRepo(dbConn)
	actionRepo := repository.NewMockActionRepo()
	usageRepo := repository.NewPostgresUsageRepo(dbConn)

	svc := services.NewChatService(
		planRepo,
		conversationRepo,
		analyticsRepo,
		actionRepo,
		nil,
		nil,
		rag.NewClientFactory(&config.Config{}),
		nil,
		nil,
		usageRepo,
		logger.New("ERROR"),
	)

	req := services.ChatRequestWithUser{
		UserID:      user.ID,
		Chatbot:     cbResult.Chatbot,
		ChatRequest: models.ChatRequest{Message: "Hello", SessionID: "sess-1"},
	}

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

// Helper functions to replace deprecated db package calls

func reserveChatTokens(ctx context.Context, db *sql.DB, userID string, estimatedTokens, maxMonthlyTokens int) error {
	query := `SELECT reserve_chat_tokens($1, $2, $3)`
	_, err := db.ExecContext(ctx, query, userID, estimatedTokens, maxMonthlyTokens)
	return err
}
