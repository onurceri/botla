package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
)

func TestJobStatusEndpoint_NoJob(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create user and get token
	token := registerAndGetToken(t, te.Server.URL, "jobtest1@example.com", fixtures.TestPassword)

	// Create chatbot
	chatbotID := createChatbot(t, te.Server.URL, token, "Job Test Bot 1")

	// Create a text source (no training job yet)
	sourceID := createTextSource(t, te.Server.URL, token, chatbotID, "Test content for job status")

	// Wait a bit for processing to potentially start
	time.Sleep(100 * time.Millisecond)

	// Get job status
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceID+"/job", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var job map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify source_id is present
	if job["source_id"] != sourceID {
		t.Errorf("expected source_id=%s, got %v", sourceID, job["source_id"])
	}

	// Verify status is returned (should be one of pending, processing, completed, failed)
	status, ok := job["status"].(string)
	if !ok {
		t.Error("expected status to be a string")
	}
	validStatuses := []string{"pending", "processing", "completed", "failed", "running"}
	found := false
	for _, s := range validStatuses {
		if status == s {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("unexpected status: %s", status)
	}

	// Verify progress_percent is returned
	progress, ok := job["progress_percent"].(float64)
	if !ok {
		t.Error("expected progress_percent to be a number")
	}
	if progress < 0 || progress > 100 {
		t.Errorf("expected progress_percent between 0-100, got %v", progress)
	}
}

func TestJobStatusEndpoint_WithJob(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create user and get token
	token := registerAndGetToken(t, te.Server.URL, "jobtest2@example.com", fixtures.TestPassword)

	// Create chatbot
	chatbotID := createChatbot(t, te.Server.URL, token, "Job Test Bot 2")

	// Create a source
	sourceID := createTextSource(t, te.Server.URL, token, chatbotID, "Content for job test")

	// Directly create a training job in the database
	trainingJobRepo := repository.NewPostgresTrainingJobRepo(te.DB)
	job, err := trainingJobRepo.Create(context.Background(), sourceID, chatbotID)
	if err != nil {
		t.Fatalf("failed to create training job: %v", err)
	}

	// Update job status to running with a step
	step := models.StepChunkText
	if err := trainingJobRepo.UpdateJobStatus(context.Background(), job.ID, models.JobStatusRunning, &step); err != nil {
		t.Fatalf("failed to update job status: %v", err)
	}

	// Get job status via API
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceID+"/job", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var jobResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify job_id is returned
	if jobResp["job_id"] != job.ID {
		t.Errorf("expected job_id=%s, got %v", job.ID, jobResp["job_id"])
	}

	// Verify status is running
	if jobResp["status"] != "running" {
		t.Errorf("expected status=running, got %v", jobResp["status"])
	}

	// Verify current_step is chunk_text
	if jobResp["current_step"] != "chunk_text" {
		t.Errorf("expected current_step=chunk_text, got %v", jobResp["current_step"])
	}

	// Verify progress_percent is 50% for chunk_text step
	if jobResp["progress_percent"] != float64(50) {
		t.Errorf("expected progress_percent=50, got %v", jobResp["progress_percent"])
	}
}

func TestJobStatusEndpoint_FailedJob(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create user and get token
	token := registerAndGetToken(t, te.Server.URL, "jobtest3@example.com", fixtures.TestPassword)

	// Create chatbot
	chatbotID := createChatbot(t, te.Server.URL, token, "Job Test Bot 3")

	// Create a source
	sourceID := createTextSource(t, te.Server.URL, token, chatbotID, "Content for failed job test")

	// Create a training job and fail it
	trainingJobRepo := repository.NewPostgresTrainingJobRepo(te.DB)
	job, err := trainingJobRepo.Create(context.Background(), sourceID, chatbotID)
	if err != nil {
		t.Fatalf("failed to create training job: %v", err)
	}

	// Fail the job
	if err := trainingJobRepo.Fail(context.Background(), job.ID, models.StepFetchSource, "FETCH_ERROR", "Could not fetch source"); err != nil {
		t.Fatalf("failed to fail job: %v", err)
	}

	// Get job status via API
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceID+"/job", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var jobResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify status is failed
	if jobResp["status"] != "failed" {
		t.Errorf("expected status=failed, got %v", jobResp["status"])
	}

	// Verify error details are returned
	if jobResp["error_code"] != "FETCH_ERROR" {
		t.Errorf("expected error_code=FETCH_ERROR, got %v", jobResp["error_code"])
	}
	if jobResp["error_message"] != "Could not fetch source" {
		t.Errorf("expected error_message='Could not fetch source', got %v", jobResp["error_message"])
	}
	if jobResp["failed_step"] != "fetch_source" {
		t.Errorf("expected failed_step=fetch_source, got %v", jobResp["failed_step"])
	}
}

func TestJobStatusEndpoint_Forbidden(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create first user and source
	token1 := registerAndGetToken(t, te.Server.URL, "owner@example.com", fixtures.TestPassword)
	chatbotID := createChatbot(t, te.Server.URL, token1, "Owner's Bot")
	sourceID := createTextSource(t, te.Server.URL, token1, chatbotID, "Owner's content")

	// Create second user (different account)
	token2 := registerAndGetToken(t, te.Server.URL, "other@example.com", fixtures.TestPassword)

	// Try to access source with different user's token
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceID+"/job", nil)
	req.Header.Set("Authorization", "Bearer "+token2)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", resp.StatusCode)
	}
}

func TestJobStatusEndpoint_NotFound(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create user and get token
	token := registerAndGetToken(t, te.Server.URL, "jobtest4@example.com", fixtures.TestPassword)

	// Try to access non-existent source
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/00000000-0000-0000-0000-000000000099/job", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", resp.StatusCode)
	}
}

func TestJobStatusEndpoint_Unauthorized(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Try to access endpoint without auth token
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/00000000-0000-0000-0000-000000000001/job", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestJobStatusEndpoint_InvalidSourceID(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create user and get token
	token := registerAndGetToken(t, te.Server.URL, "jobtest5@example.com", fixtures.TestPassword)

	// Try to access with invalid UUID format
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/invalid-uuid/job", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", resp.StatusCode)
	}
}

func TestJobStatusEndpoint_CompletedJob(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create user and get token
	token := registerAndGetToken(t, te.Server.URL, "jobtest6@example.com", fixtures.TestPassword)

	// Create chatbot
	chatbotID := createChatbot(t, te.Server.URL, token, "Job Test Bot 6")

	// Create a source
	sourceID := createTextSource(t, te.Server.URL, token, chatbotID, "Content for completed job test")

	// Create a training job and complete it
	trainingJobRepo := repository.NewPostgresTrainingJobRepo(te.DB)
	job, err := trainingJobRepo.Create(context.Background(), sourceID, chatbotID)
	if err != nil {
		t.Fatalf("failed to create training job: %v", err)
	}

	// Complete the job
	if err := trainingJobRepo.Complete(context.Background(), job.ID); err != nil {
		t.Fatalf("failed to complete job: %v", err)
	}

	// Get job status via API
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceID+"/job", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var jobResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify status is completed
	if jobResp["status"] != "completed" {
		t.Errorf("expected status=completed, got %v", jobResp["status"])
	}

	// Verify progress is 100%
	if jobResp["progress_percent"] != float64(100) {
		t.Errorf("expected progress_percent=100, got %v", jobResp["progress_percent"])
	}

	// Verify completed_at is set
	if jobResp["completed_at"] == nil {
		t.Error("expected completed_at to be set")
	}
}

// Helper function to create a chatbot
func createChatbot(t *testing.T, baseURL, token, name string) string {
	t.Helper()

	payload := map[string]string{"name": name}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := testHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create chatbot: %v", err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create chatbot, status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode chatbot response: %v", err)
	}

	id, ok := result["id"].(string)
	if !ok {
		t.Fatal("chatbot id not found in response")
	}

	return id
}

// Helper function to create a text source
func createTextSource(t *testing.T, baseURL, token, chatbotID, content string) string {
	t.Helper()

	// Use multipart form data like the API expects
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "text")
	mw.WriteField("text", content)
	mw.Close()

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots/"+chatbotID+"/sources", &b)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	client := testHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}
	defer drainBody(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create source, status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode source response: %v", err)
	}

	id, ok := result["id"].(string)
	if !ok {
		t.Fatal("source id not found in response")
	}

	return id
}

// Helper function to register a user and get a token
func registerAndGetToken(t *testing.T, baseURL, email, password string) string {
	t.Helper()

	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	// Register
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := testHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	defer drainBody(resp)

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}

	// Try both "token" and "access_token" for compatibility
	token, ok := result["token"].(string)
	if !ok {
		token, ok = result["access_token"].(string)
		if !ok {
			t.Fatalf("token not found in response: %v", result)
		}
	}

	return token
}
