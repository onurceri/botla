package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/onurceri/botla-co/internal/models"
)

func setupTestDBForSuggestionJob(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("postgres", "postgres://botla:botla@localhost/botla_test?sslmode=disable&options=-c%20search_path%3Dtest")
	if err != nil {
		t.Skipf("skipping test: could not connect to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Skipf("skipping test: could not ping database: %v", err)
	}

	teardown := func() {
		db.Close()
	}

	return db, teardown
}

func createTestChatbotForJobTest(t *testing.T, db *sql.DB) string {
	ctx := context.Background()
	chatbotID := uuid.NewString()
	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	orgID := uuid.NewString()

	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, plan_id, onboarding_completed, onboarding_step, onboarding_skipped)
		VALUES ($1, $2, $3, $4, (SELECT id FROM plans LIMIT 1), true, 0, false)
	`, userID, uuid.NewString()+"@test.com", "hashed", "Test User")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO organizations (id, name, slug, owner_id, plan_id)
		VALUES ($1, $2, $3, $4, 'agency_starter')
	`, orgID, "Test Org", "test-org-"+uuid.NewString()[:8], userID)
	if err != nil {
		t.Fatalf("failed to create test org: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO workspaces (id, organization_id, name, slug)
		VALUES ($1, $2, $3, $4)
	`, workspaceID, orgID, "Test Workspace", "test-ws-"+uuid.NewString()[:8])
	if err != nil {
		t.Fatalf("failed to create test workspace: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO chatbots (id, name, user_id, workspace_id, organization_id)
		VALUES ($1, $2, $3, $4, $5)
	`, chatbotID, "test-bot-"+uuid.NewString()[:8], userID, workspaceID, orgID)
	if err != nil {
		t.Fatalf("failed to create test chatbot: %v", err)
	}

	return chatbotID
}

func TestCreateSuggestionJob(t *testing.T) {
	db, teardown := setupTestDBForSuggestionJob(t)
	defer teardown()

	ctx := context.Background()
	chatbotID := createTestChatbotForJobTest(t, db)

	job, err := CreateSuggestionJob(ctx, db, chatbotID)
	if err != nil {
		t.Fatalf("CreateSuggestionJob failed: %v", err)
	}

	if job.ID == "" {
		t.Errorf("expected job ID to be set")
	}

	if job.ChatbotID != chatbotID {
		t.Errorf("expected chatbot ID %s, got %s", chatbotID, job.ChatbotID)
	}

	if job.Status != models.SuggestionJobStatusPending {
		t.Errorf("expected status pending, got %s", job.Status)
	}

	if job.CreatedAt.IsZero() {
		t.Errorf("expected created_at to be set")
	}
}

func TestGetSuggestionJob(t *testing.T) {
	db, teardown := setupTestDBForSuggestionJob(t)
	defer teardown()

	ctx := context.Background()
	chatbotID := createTestChatbotForJobTest(t, db)

	createdJob, err := CreateSuggestionJob(ctx, db, chatbotID)
	if err != nil {
		t.Fatalf("CreateSuggestionJob failed: %v", err)
	}

	job, err := GetSuggestionJob(ctx, db, createdJob.ID)
	if err != nil {
		t.Fatalf("GetSuggestionJob failed: %v", err)
	}

	if job == nil {
		t.Fatalf("expected job, got nil")
	}

	if job.ID != createdJob.ID {
		t.Errorf("expected job ID %s, got %s", createdJob.ID, job.ID)
	}

	if job.Status != models.SuggestionJobStatusPending {
		t.Errorf("expected status pending, got %s", job.Status)
	}
}

func TestGetSuggestionJob_NotFound(t *testing.T) {
	db, teardown := setupTestDBForSuggestionJob(t)
	defer teardown()

	ctx := context.Background()

	job, err := GetSuggestionJob(ctx, db, uuid.NewString())
	if err != nil {
		t.Fatalf("GetSuggestionJob failed: %v", err)
	}

	if job != nil {
		t.Errorf("expected nil job, got %v", job)
	}
}

func TestUpdateSuggestionJobStatus(t *testing.T) {
	db, teardown := setupTestDBForSuggestionJob(t)
	defer teardown()

	ctx := context.Background()
	chatbotID := createTestChatbotForJobTest(t, db)

	job, err := CreateSuggestionJob(ctx, db, chatbotID)
	if err != nil {
		t.Fatalf("CreateSuggestionJob failed: %v", err)
	}

	err = UpdateSuggestionJobStatus(ctx, db, job.ID, models.SuggestionJobStatusRunning)
	if err != nil {
		t.Fatalf("UpdateSuggestionJobStatus failed: %v", err)
	}

	updatedJob, err := GetSuggestionJob(ctx, db, job.ID)
	if err != nil {
		t.Fatalf("GetSuggestionJob failed: %v", err)
	}

	if updatedJob.Status != models.SuggestionJobStatusRunning {
		t.Errorf("expected status running, got %s", updatedJob.Status)
	}

	if updatedJob.StartedAt == nil {
		t.Errorf("expected started_at to be set")
	}
}

func TestCompleteSuggestionJob(t *testing.T) {
	db, teardown := setupTestDBForSuggestionJob(t)
	defer teardown()

	ctx := context.Background()
	chatbotID := createTestChatbotForJobTest(t, db)

	job, err := CreateSuggestionJob(ctx, db, chatbotID)
	if err != nil {
		t.Fatalf("CreateSuggestionJob failed: %v", err)
	}

	suggestions := []string{"Q1", "Q2", "Q3"}
	err = CompleteSuggestionJob(ctx, db, job.ID, suggestions)
	if err != nil {
		t.Fatalf("CompleteSuggestionJob failed: %v", err)
	}

	completedJob, err := GetSuggestionJob(ctx, db, job.ID)
	if err != nil {
		t.Fatalf("GetSuggestionJob failed: %v", err)
	}

	if completedJob.Status != models.SuggestionJobStatusCompleted {
		t.Errorf("expected status completed, got %s", completedJob.Status)
	}

	if len(completedJob.SuggestedQuestions) != 3 {
		t.Errorf("expected 3 suggestions, got %d", len(completedJob.SuggestedQuestions))
	}

	if completedJob.CompletedAt == nil {
		t.Errorf("expected completed_at to be set")
	}
}

func TestFailSuggestionJob(t *testing.T) {
	db, teardown := setupTestDBForSuggestionJob(t)
	defer teardown()

	ctx := context.Background()
	chatbotID := createTestChatbotForJobTest(t, db)

	job, err := CreateSuggestionJob(ctx, db, chatbotID)
	if err != nil {
		t.Fatalf("CreateSuggestionJob failed: %v", err)
	}

	errMsg := "test error message"
	err = FailSuggestionJob(ctx, db, job.ID, errMsg)
	if err != nil {
		t.Fatalf("FailSuggestionJob failed: %v", err)
	}

	failedJob, err := GetSuggestionJob(ctx, db, job.ID)
	if err != nil {
		t.Fatalf("GetSuggestionJob failed: %v", err)
	}

	if failedJob.Status != models.SuggestionJobStatusFailed {
		t.Errorf("expected status failed, got %s", failedJob.Status)
	}

	if failedJob.ErrorMessage == nil || *failedJob.ErrorMessage != errMsg {
		t.Errorf("expected error message %q, got %v", errMsg, failedJob.ErrorMessage)
	}

	if failedJob.CompletedAt == nil {
		t.Errorf("expected completed_at to be set")
	}
}

func TestGetLatestSuggestionJobForChatbot(t *testing.T) {
	db, teardown := setupTestDBForSuggestionJob(t)
	defer teardown()

	ctx := context.Background()
	chatbotID := createTestChatbotForJobTest(t, db)

	job1, err := CreateSuggestionJob(ctx, db, chatbotID)
	if err != nil {
		t.Fatalf("CreateSuggestionJob failed: %v", err)
	}
	_ = job1

	time.Sleep(10 * time.Millisecond)

	job2, err := CreateSuggestionJob(ctx, db, chatbotID)
	if err != nil {
		t.Fatalf("CreateSuggestionJob failed: %v", err)
	}

	latestJob, err := GetLatestSuggestionJobForChatbot(ctx, db, chatbotID)
	if err != nil {
		t.Fatalf("GetLatestSuggestionJobForChatbot failed: %v", err)
	}

	if latestJob == nil {
		t.Fatalf("expected job, got nil")
	}

	if latestJob.ID != job2.ID {
		t.Errorf("expected latest job ID %s, got %s", job2.ID, latestJob.ID)
	}
}
