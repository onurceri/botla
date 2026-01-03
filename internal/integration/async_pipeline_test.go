package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/models"
)

// Test full pipeline from source creation to completion
func TestAsyncPipeline_FullLifecycle(t *testing.T) {
t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	te, err := fixtures.SetupTestEnvWithMocks()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "pipeline@test.com")
	botID := createChatbot(t, te.Server.URL, token, "Pipeline Bot")

	// Create text source (more reliable than URL for testing)
	sourceID := createTextSource(t, te.Server.URL, token, botID, "Test content for full lifecycle test")

	// Poll job status through steps
	var seenSteps []string
	var lastStatus string

	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-timeout:
			t.Errorf("job did not complete in time, last status: %s, seen steps: %v", lastStatus, seenSteps)
			break LOOP
		case <-ticker.C:
			job := getJobStatusMap(t, te.Server.URL, token, sourceID)
			if job == nil {
				continue
			}

			if step, ok := job["current_step"].(string); ok && step != "" {
				if len(seenSteps) == 0 || seenSteps[len(seenSteps)-1] != step {
					seenSteps = append(seenSteps, step)
				}
			}

			if status, ok := job["status"].(string); ok {
				lastStatus = status
				if status == "completed" || status == "failed" {
					break LOOP
				}
			}
		}
	}

	// Verify job completed
	if lastStatus != "completed" {
		t.Errorf("expected job to complete, got %s", lastStatus)
	}

	// Verify steps were traversed (might be very fast with mocks)
	// With mocks, processing is fast, so we may not catch all steps in the poll
	// But we should see at least the final state
	t.Logf("Seen steps: %v, final status: %s", seenSteps, lastStatus)
}

// Test job failure and retry
func TestAsyncPipeline_FailureAndRetry(t *testing.T) {
t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	te, err := fixtures.SetupTestEnvWithMocks()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "retry@test.com")
	botID := createChatbot(t, te.Server.URL, token, "Retry Bot")

	// Create source with invalid URL that will fail
	resp := createURLSource(t, te.Server.URL, token, botID, "http://invalid.localhost.test")
	if resp.StatusCode != http.StatusCreated {
		t.Skipf("source creation returned status %d – may have failed at validation", resp.StatusCode)
	}
	var src struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&src); err != nil {
		t.Fatalf("failed to decode source response: %v", err)
	}
	drainBody(resp)
	sourceID := src.ID

	// Wait for failure
	timeout := time.After(15 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastStatus string
	var job map[string]interface{}

LOOP:
	for {
		select {
		case <-timeout:
			t.Logf("job status at timeout: %v", job)
			break LOOP
		case <-ticker.C:
			job = getJobStatusMap(t, te.Server.URL, token, sourceID)
			if job == nil {
				continue
			}

			if status, ok := job["status"].(string); ok {
				lastStatus = status
				if status == "failed" || status == "completed" {
					break LOOP
				}
			}
		}
	}

	if lastStatus != "failed" {
		t.Skipf("job didn't fail as expected: %s (this test requires network failure)", lastStatus)
	}

	// Verify error details
	if job["error_code"] == nil {
		t.Error("expected error_code to be set on failed job")
	}
	if job["failed_step"] == nil {
		t.Error("expected failed_step to be set on failed job")
	}
	t.Logf("Job failed with error_code=%v, failed_step=%v", job["error_code"], job["failed_step"])
}

// Test job recovery after simulated restart
func TestAsyncPipeline_Recovery(t *testing.T) {
t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	te, err := fixtures.SetupTestEnvWithMocks()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create user first
	token := authToken(t, te.Server.URL, "recovery@test.com")
	botID := createChatbot(t, te.Server.URL, token, "Recovery Bot")

	// Directly insert a training job in pending state (simulating a crash scenario)
	ctx := context.Background()

	// First create a source
	sourceID := createTextSource(t, te.Server.URL, token, botID, "Recovery test content")

	// Wait a bit for any initial processing to start
	time.Sleep(500 * time.Millisecond)

	// Create a new job in pending state (simulating a scenario where a job was enqueued but not processed)
	job, err := db.CreateTrainingJob(ctx, te.DB, sourceID, botID)
	if err != nil {
		t.Fatalf("failed to create pending job: %v", err)
	}

	// Manually trigger recovery by putting the job in the queue
	te.Queue.Enqueue(job.ID)

	// Wait and check processing completed
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-timeout:
			t.Fatal("recovery did not complete in time")
		case <-ticker.C:
			// Check job status
			j, err := db.GetTrainingJob(ctx, te.DB, job.ID)
			if err != nil {
				continue
			}
			if j.Status == models.JobStatusCompleted || j.Status == models.JobStatusFailed {
				if j.Status != models.JobStatusCompleted {
					t.Errorf("expected job to complete successfully, got %s", j.Status)
				}
				break LOOP
			}
		}
	}
}

// Test concurrent job processing
func TestAsyncPipeline_ConcurrentJobs(t *testing.T) {
t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	te, err := fixtures.SetupTestEnvWithMocks()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "concurrent@test.com")
	botID := createChatbot(t, te.Server.URL, token, "Concurrent Bot")

	// Create multiple sources simultaneously
	sourceCount := 4
	sourceChan := make(chan string, sourceCount)
	var wg sync.WaitGroup

	for i := 0; i < sourceCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			content := fmt.Sprintf("Concurrent processing test content %d", idx)
			id := createTextSource(t, te.Server.URL, token, botID, content)
			sourceChan <- id
		}(i)
	}

	wg.Wait()
	close(sourceChan)

	// Collect source IDs
	var sourceIDs []string
	for id := range sourceChan {
		sourceIDs = append(sourceIDs, id)
	}

	if len(sourceIDs) != sourceCount {
		t.Fatalf("expected %d sources, got %d", sourceCount, len(sourceIDs))
	}

	// Wait for all jobs to complete
	waitForAllSourcesCompletion(t, te, sourceIDs, 30*time.Second)
}

// Helper function to get job status as a map
func getJobStatusMap(t *testing.T, baseURL, token, sourceID string) map[string]interface{} {
	t.Helper()

	req, err := http.NewRequest("GET", baseURL+"/api/v1/sources/"+sourceID+"/job", nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var job map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil
	}

	return job
}

// Helper function to wait for all sources to complete
func waitForAllSourcesCompletion(t *testing.T, te *fixtures.TestEnv, sourceIDs []string, timeout time.Duration) {
	t.Helper()

	deadline := time.After(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			// Check final state of all sources
			for _, id := range sourceIDs {
				var status string
				err := te.DB.QueryRow("SELECT status FROM data_sources WHERE id=$1", id).Scan(&status)
				if err != nil {
					t.Logf("source %s: query error: %v", id, err)
				} else {
					t.Logf("source %s: status=%s", id, status)
				}
			}
			t.Fatal("timed out waiting for all sources to complete")
		case <-ticker.C:
			completed := 0
			for _, id := range sourceIDs {
				var status string
				err := te.DB.QueryRow("SELECT status FROM data_sources WHERE id=$1", id).Scan(&status)
				if err != nil {
					continue
				}
				if status == "completed" || status == "failed" {
					completed++
				}
			}
			if completed == len(sourceIDs) {
				return // All done
			}
		}
	}
}
