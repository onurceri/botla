package db_test

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
)

func TestCreateTrainingJob(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	
	// Create test source with full hierarchy
	sourceResult := testdb.CreateSource(t, dbConn)
	
	job, err := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	if err != nil {
		t.Fatalf("CreateTrainingJob failed: %v", err)
	}

	if job.ID == "" {
		t.Error("expected job ID, got empty")
	}
	if job.SourceID != sourceResult.Source.ID {
		t.Errorf("expected source ID %s, got %s", sourceResult.Source.ID, job.SourceID)
	}
	if job.ChatbotID != sourceResult.Chatbot.ID {
		t.Errorf("expected chatbot ID %s, got %s", sourceResult.Chatbot.ID, job.ChatbotID)
	}
	if job.Status != models.JobStatusPending {
		t.Errorf("expected pending status, got %s", job.Status)
	}
	if job.ProgressPercent != 0 {
		t.Errorf("expected 0 progress, got %d", job.ProgressPercent)
	}
	if job.RetryCount != 0 {
		t.Errorf("expected 0 retry count, got %d", job.RetryCount)
	}
}

func TestGetTrainingJob(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	job, err := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	if err != nil {
		t.Fatalf("CreateTrainingJob failed: %v", err)
	}

	// Retrieve the job
	retrieved, err := db.GetTrainingJob(context.Background(), dbConn, job.ID)
	if err != nil {
		t.Fatalf("GetTrainingJob failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected job, got nil")
	}
	if retrieved.ID != job.ID {
		t.Errorf("expected ID %s, got %s", job.ID, retrieved.ID)
	}
	if retrieved.Status != models.JobStatusPending {
		t.Errorf("expected pending, got %s", retrieved.Status)
	}

	// Test non-existent job
	notFound, err := db.GetTrainingJob(context.Background(), dbConn, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("GetTrainingJob for non-existent ID failed: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent job")
	}
}

func TestUpdateJobStatus(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	job, err := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	if err != nil {
		t.Fatalf("CreateTrainingJob failed: %v", err)
	}

	// Update to running with fetch_source step
	step := models.StepFetchSource
	err = db.UpdateJobStatus(context.Background(), dbConn, job.ID, models.JobStatusRunning, &step)
	if err != nil {
		t.Fatalf("UpdateJobStatus failed: %v", err)
	}

	// Verify
	updated, err := db.GetTrainingJob(context.Background(), dbConn, job.ID)
	if err != nil {
		t.Fatalf("GetTrainingJob failed: %v", err)
	}
	if updated.Status != models.JobStatusRunning {
		t.Errorf("expected running, got %s", updated.Status)
	}
	if updated.CurrentStep == nil || *updated.CurrentStep != models.StepFetchSource {
		t.Errorf("expected fetch_source step, got %v", updated.CurrentStep)
	}
	if updated.ProgressPercent != 10 {
		t.Errorf("expected 10 progress (fetch_source), got %d", updated.ProgressPercent)
	}
	if updated.StartedAt == nil {
		t.Error("expected started_at to be set")
	}

	// Update to embed_chunks step
	embedStep := models.StepEmbedChunks
	err = db.UpdateJobStatus(context.Background(), dbConn, job.ID, models.JobStatusRunning, &embedStep)
	if err != nil {
		t.Fatalf("UpdateJobStatus to embed_chunks failed: %v", err)
	}

	updated2, _ := db.GetTrainingJob(context.Background(), dbConn, job.ID)
	if updated2.ProgressPercent != 80 {
		t.Errorf("expected 80 progress (embed_chunks), got %d", updated2.ProgressPercent)
	}
}

func TestFailJob(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	job, err := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	if err != nil {
		t.Fatalf("CreateTrainingJob failed: %v", err)
	}

	// Start the job first
	step := models.StepEmbedChunks
	_ = db.UpdateJobStatus(context.Background(), dbConn, job.ID, models.JobStatusRunning, &step)

	// Fail it
	err = db.FailJob(context.Background(), dbConn, job.ID, models.StepEmbedChunks, "EMBED_ERROR", "OpenAI API rate limit")
	if err != nil {
		t.Fatalf("FailJob failed: %v", err)
	}

	updated, err := db.GetTrainingJob(context.Background(), dbConn, job.ID)
	if err != nil {
		t.Fatalf("GetTrainingJob failed: %v", err)
	}
	if updated.Status != models.JobStatusFailed {
		t.Errorf("expected failed, got %s", updated.Status)
	}
	if updated.FailedStep == nil || *updated.FailedStep != models.StepEmbedChunks {
		t.Errorf("expected embed_chunks failed step, got %v", updated.FailedStep)
	}
	if updated.ErrorCode == nil || *updated.ErrorCode != "EMBED_ERROR" {
		t.Errorf("expected EMBED_ERROR, got %v", updated.ErrorCode)
	}
	if updated.ErrorMessage == nil || *updated.ErrorMessage != "OpenAI API rate limit" {
		t.Errorf("expected 'OpenAI API rate limit', got %v", updated.ErrorMessage)
	}
	if updated.CompletedAt == nil {
		t.Error("expected completed_at to be set on failure")
	}
}

func TestCompleteJob(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	job, err := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	if err != nil {
		t.Fatalf("CreateTrainingJob failed: %v", err)
	}

	// Complete the job
	err = db.CompleteJob(context.Background(), dbConn, job.ID)
	if err != nil {
		t.Fatalf("CompleteJob failed: %v", err)
	}

	updated, _ := db.GetTrainingJob(context.Background(), dbConn, job.ID)
	if updated.Status != models.JobStatusCompleted {
		t.Errorf("expected completed, got %s", updated.Status)
	}
	if updated.ProgressPercent != 100 {
		t.Errorf("expected 100 progress, got %d", updated.ProgressPercent)
	}
	if updated.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestCancelJob(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	job, err := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	if err != nil {
		t.Fatalf("CreateTrainingJob failed: %v", err)
	}

	// Cancel the job
	err = db.CancelJob(context.Background(), dbConn, job.ID)
	if err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}

	updated, _ := db.GetTrainingJob(context.Background(), dbConn, job.ID)
	if updated.Status != models.JobStatusCancelled {
		t.Errorf("expected cancelled, got %s", updated.Status)
	}
	if updated.CompletedAt == nil {
		t.Error("expected completed_at to be set on cancellation")
	}
}

func TestGetJobBySourceID(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	// Create multiple jobs for the same source
	job1, _ := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	_ = db.CompleteJob(context.Background(), dbConn, job1.ID)
	
	job2, _ := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)

	// Should return the most recent job (job2)
	latest, err := db.GetJobBySourceID(context.Background(), dbConn, sourceResult.Source.ID)
	if err != nil {
		t.Fatalf("GetJobBySourceID failed: %v", err)
	}
	if latest == nil {
		t.Fatal("expected job, got nil")
	}
	if latest.ID != job2.ID {
		t.Errorf("expected job2 ID %s, got %s", job2.ID, latest.ID)
	}

	// Test non-existent source
	notFound, err := db.GetJobBySourceID(context.Background(), dbConn, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("GetJobBySourceID for non-existent source failed: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent source")
	}
}

func TestGetJobsByChatbotID(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	
	// Create two sources for the same chatbot
	sourceResult1 := testdb.CreateSource(t, dbConn)
	sourceResult2 := testdb.CreateSource(t, dbConn, testdb.SourceFixture{
		ChatbotID: sourceResult1.Chatbot.ID,
	})

	// Create jobs
	_, _ = db.CreateTrainingJob(context.Background(), dbConn, sourceResult1.Source.ID, sourceResult1.Chatbot.ID)
	_, _ = db.CreateTrainingJob(context.Background(), dbConn, sourceResult2.Source.ID, sourceResult1.Chatbot.ID)

	jobs, err := db.GetJobsByChatbotID(context.Background(), dbConn, sourceResult1.Chatbot.ID, 10)
	if err != nil {
		t.Fatalf("GetJobsByChatbotID failed: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestGetPendingJobs(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	// Create pending jobs
	job1, _ := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	job2, _ := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	
	// Make job1 running
	step := models.StepFetchSource
	_ = db.UpdateJobStatus(context.Background(), dbConn, job1.ID, models.JobStatusRunning, &step)

	// Get pending jobs
	pending, err := db.GetPendingJobs(context.Background(), dbConn, 100)
	if err != nil {
		t.Fatalf("GetPendingJobs failed: %v", err)
	}
	
	// Verify at least our pending job (job2) is returned and running job (job1) is NOT
	foundJob2 := false
	foundJob1 := false
	for _, job := range pending {
		if job.ID == job2.ID {
			foundJob2 = true
			if job.Status != models.JobStatusPending {
				t.Errorf("job2 should have pending status, got %s", job.Status)
			}
		}
		if job.ID == job1.ID {
			foundJob1 = true
		}
	}
	
	if !foundJob2 {
		t.Errorf("job2 (ID=%s) should be in pending list, got %d pending jobs", job2.ID, len(pending))
	}
	if foundJob1 {
		t.Errorf("job1 (ID=%s) should not be in pending list (it's running)", job1.ID)
	}
}

func TestGetRetryableJobs(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	job, _ := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	
	// Fail the job
	_ = db.FailJob(context.Background(), dbConn, job.ID, models.StepEmbedChunks, "ERROR", "test error")

	// Should be retryable with maxRetries=3
	retryable, err := db.GetRetryableJobs(context.Background(), dbConn, 3, 10)
	if err != nil {
		t.Fatalf("GetRetryableJobs failed: %v", err)
	}
	
	found := false
	for _, j := range retryable {
		if j.ID == job.ID {
			found = true
		}
	}
	if !found {
		t.Error("failed job should be retryable")
	}
}

func TestIncrementJobRetryCount(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	job, _ := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	
	// Fail the job
	_ = db.FailJob(context.Background(), dbConn, job.ID, models.StepEmbedChunks, "ERROR", "test error")

	// Increment retry count
	err := db.IncrementJobRetryCount(context.Background(), dbConn, job.ID)
	if err != nil {
		t.Fatalf("IncrementJobRetryCount failed: %v", err)
	}

	updated, _ := db.GetTrainingJob(context.Background(), dbConn, job.ID)
	if updated.RetryCount != 1 {
		t.Errorf("expected retry count 1, got %d", updated.RetryCount)
	}
	if updated.Status != models.JobStatusPending {
		t.Errorf("expected pending status after retry reset, got %s", updated.Status)
	}
	if updated.ErrorCode != nil {
		t.Error("expected error_code to be cleared")
	}
	if updated.ErrorMessage != nil {
		t.Error("expected error_message to be cleared")
	}
}

func TestGetRunningJobs(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	sourceResult := testdb.CreateSource(t, dbConn)
	
	job, _ := db.CreateTrainingJob(context.Background(), dbConn, sourceResult.Source.ID, sourceResult.Chatbot.ID)
	
	// Make job running
	step := models.StepParseContent
	_ = db.UpdateJobStatus(context.Background(), dbConn, job.ID, models.JobStatusRunning, &step)

	// Retrieve with a larger limit to account for other running jobs from other tests
	running, err := db.GetRunningJobs(context.Background(), dbConn, 100)
	if err != nil {
		t.Fatalf("GetRunningJobs failed: %v", err)
	}
	
	found := false
	for _, j := range running {
		if j.ID == job.ID {
			found = true
			if j.CurrentStep == nil || *j.CurrentStep != models.StepParseContent {
				t.Errorf("expected parse_content step, got %v", j.CurrentStep)
			}
			if j.Status != models.JobStatusRunning {
				t.Errorf("expected running status, got %s", j.Status)
			}
			break
		}
	}
	if !found {
		t.Errorf("running job (ID=%s) should be in list, got %d running jobs", job.ID, len(running))
	}
}

func TestTrainingJobModel(t *testing.T) {
	t.Run("IsTerminal", func(t *testing.T) {
		job := &models.TrainingJob{Status: models.JobStatusPending}
		if job.IsTerminal() {
			t.Error("pending should not be terminal")
		}

		job.Status = models.JobStatusRunning
		if job.IsTerminal() {
			t.Error("running should not be terminal")
		}

		job.Status = models.JobStatusCompleted
		if !job.IsTerminal() {
			t.Error("completed should be terminal")
		}

		job.Status = models.JobStatusFailed
		if !job.IsTerminal() {
			t.Error("failed should be terminal")
		}

		job.Status = models.JobStatusCancelled
		if !job.IsTerminal() {
			t.Error("cancelled should be terminal")
		}
	})

	t.Run("CanRetry", func(t *testing.T) {
		job := &models.TrainingJob{
			Status:     models.JobStatusFailed,
			RetryCount: 0,
		}
		if !job.CanRetry(3) {
			t.Error("failed job with 0 retries should be retryable")
		}

		job.RetryCount = 3
		if job.CanRetry(3) {
			t.Error("failed job with 3 retries should not be retryable when max is 3")
		}

		job.Status = models.JobStatusPending
		job.RetryCount = 0
		if job.CanRetry(3) {
			t.Error("pending job should not be retryable")
		}
	})

	t.Run("JobStatusIsValid", func(t *testing.T) {
		if !models.JobStatusPending.IsValid() {
			t.Error("pending should be valid")
		}
		if models.JobStatus("invalid").IsValid() {
			t.Error("invalid status should not be valid")
		}
	})

	t.Run("TrainingStepIsValid", func(t *testing.T) {
		if !models.StepFetchSource.IsValid() {
			t.Error("fetch_source should be valid")
		}
		if models.TrainingStep("invalid").IsValid() {
			t.Error("invalid step should not be valid")
		}
	})

	t.Run("StepProgress", func(t *testing.T) {
		if models.StepProgress[models.StepFetchSource] != 10 {
			t.Errorf("fetch_source progress should be 10, got %d", models.StepProgress[models.StepFetchSource])
		}
		if models.StepProgress[models.StepStoreVectors] != 100 {
			t.Errorf("store_vectors progress should be 100, got %d", models.StepProgress[models.StepStoreVectors])
		}
	})
}
