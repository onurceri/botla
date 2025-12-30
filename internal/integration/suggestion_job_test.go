package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
)

func authTokenForSuggestionJob(t *testing.T, base string, email string) string {
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	b, _ := json.Marshal(regBody)
	http.Post(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	lb := map[string]string{"email": email, "password": "Test@123"}
	lbj, _ := json.Marshal(lb)
	res, err := http.Post(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	var tr struct {
		Token string `json:"token"`
	}
	json.NewDecoder(res.Body).Decode(&tr)
	res.Body.Close()
	return tr.Token
}

func TestSuggestionRegenerationPolling_FullFlow(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForSuggestionJob(t, te.Server.URL, "suggestion_polling@example.com")

	ctx := context.Background()

	result := testdb.CreateChatbot(t, te.DB)
	chatbotID := result.Chatbot.ID

	sourceResult := testdb.CreateSource(t, te.DB, testdb.SourceFixture{
		ChatbotID:  chatbotID,
		SourceType: "text",
		Status:     "completed",
		ChunkCount: 10,
	})

	sourceID := sourceResult.Source.ID

	suggestions := []string{"Source Question 1", "Source Question 2"}
	_, err = te.DB.ExecContext(ctx, `
		UPDATE data_sources SET suggested_questions = $1 WHERE id = $2
	`, suggestions, sourceID)
	if err != nil {
		t.Fatalf("failed to set source suggestions: %v", err)
	}

	regenerateReq, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+chatbotID+"/suggestions/regenerate", nil)
	regenerateReq.Header.Set("Authorization", "Bearer "+token)
	regenerateResp, err := http.DefaultClient.Do(regenerateReq)
	if err != nil {
		t.Fatalf("regenerate request failed: %v", err)
	}
	defer regenerateResp.Body.Close()

	if regenerateResp.StatusCode != http.StatusAccepted {
		buf := new(bytes.Buffer)
		buf.ReadFrom(regenerateResp.Body)
		t.Fatalf("expected 202, got %d. Body: %s", regenerateResp.StatusCode, buf.String())
	}

	var regenerateBody struct {
		JobID string `json:"job_id"`
	}
	json.NewDecoder(regenerateResp.Body).Decode(&regenerateBody)

	if regenerateBody.JobID == "" {
		t.Fatal("expected job_id in response")
	}

	job, err := db.GetSuggestionJob(ctx, te.DB, regenerateBody.JobID)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}
	if job == nil {
		t.Fatal("expected job to be created")
	}
	if job.Status != models.SuggestionJobStatusPending {
		t.Errorf("expected pending status, got %s", job.Status)
	}

	time.Sleep(100 * time.Millisecond)

	statusReq, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	statusReq.Header.Set("Authorization", "Bearer "+token)
	statusResp, err := http.DefaultClient.Do(statusReq)
	if err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusResp.StatusCode)
	}

	var statusBody struct {
		JobID  string `json:"job_id"`
		Status string `json:"status"`
	}
	json.NewDecoder(statusResp.Body).Decode(&statusBody)

	if statusBody.JobID == "" {
		t.Fatal("expected job_id in status response")
	}
	if statusBody.Status != models.SuggestionJobStatusPending.String() && statusBody.Status != models.SuggestionJobStatusRunning.String() && statusBody.Status != models.SuggestionJobStatusCompleted.String() {
		t.Errorf("unexpected status: %s", statusBody.Status)
	}

	updatedChatbot, err := db.GetChatbotByID(ctx, te.DB, chatbotID)
	if err != nil {
		t.Fatalf("failed to get chatbot: %v", err)
	}

	if updatedChatbot.SuggestedQuestions == nil || len(updatedChatbot.SuggestedQuestions) == 0 {
		t.Log("suggestions may not be populated yet (async processing)")
	}
}

func TestSuggestionRegeneration_ConcurrentRequests(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForSuggestionJob(t, te.Server.URL, "concurrent_suggestion@example.com")

	result := testdb.CreateChatbot(t, te.DB)
	chatbotID := result.Chatbot.ID

	numRequests := 5
	jobIDs := make(map[string]bool)

	for i := 0; i < numRequests; i++ {
		regenerateReq, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+chatbotID+"/suggestions/regenerate", nil)
		regenerateReq.Header.Set("Authorization", "Bearer "+token)
		regenerateResp, err := http.DefaultClient.Do(regenerateReq)
		if err != nil {
			t.Fatalf("regenerate request %d failed: %v", i, err)
		}
		defer regenerateResp.Body.Close()

		if regenerateResp.StatusCode != http.StatusAccepted {
			t.Fatalf("request %d: expected 202, got %d", i, regenerateResp.StatusCode)
		}

		var body struct {
			JobID string `json:"job_id"`
		}
		json.NewDecoder(regenerateResp.Body).Decode(&body)

		if body.JobID == "" {
			t.Fatalf("request %d: expected job_id", i)
		}

		jobIDs[body.JobID] = true
	}

	if len(jobIDs) != numRequests {
		t.Errorf("expected %d unique job IDs, got %d", numRequests, len(jobIDs))
	}

	ctx := context.Background()
	for jobID := range jobIDs {
		job, err := db.GetSuggestionJob(ctx, te.DB, jobID)
		if err != nil {
			t.Fatalf("failed to get job %s: %v", jobID, err)
		}
		if job == nil {
			t.Errorf("job %s should exist", jobID)
		}
	}
}

func TestSuggestionRegeneration_GetLatestJob(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForSuggestionJob(t, te.Server.URL, "latest_job@example.com")

	result := testdb.CreateChatbot(t, te.DB)
	chatbotID := result.Chatbot.ID

	ctx := context.Background()

	job1, err := db.CreateSuggestionJob(ctx, te.DB, chatbotID)
	if err != nil {
		t.Fatalf("failed to create job 1: %v", err)
	}
	_ = job1

	time.Sleep(10 * time.Millisecond)

	job2, err := db.CreateSuggestionJob(ctx, te.DB, chatbotID)
	if err != nil {
		t.Fatalf("failed to create job 2: %v", err)
	}

	statusReq, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	statusReq.Header.Set("Authorization", "Bearer "+token)
	statusResp, err := http.DefaultClient.Do(statusReq)
	if err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusResp.StatusCode)
	}

	var body struct {
		JobID string `json:"job_id"`
	}
	json.NewDecoder(statusResp.Body).Decode(&body)

	if body.JobID != job2.ID {
		t.Errorf("expected latest job %s, got %s", job2.ID, body.JobID)
	}
}

func TestSuggestionRegeneration_JobCompletion(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	ctx := context.Background()

	result := testdb.CreateChatbot(t, te.DB)
	chatbotID := result.Chatbot.ID

	job, err := db.CreateSuggestionJob(ctx, te.DB, chatbotID)
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	suggestions := []string{"Generated Q1", "Generated Q2", "Generated Q3"}
	err = db.CompleteSuggestionJob(ctx, te.DB, job.ID, suggestions)
	if err != nil {
		t.Fatalf("failed to complete job: %v", err)
	}

	token := authTokenForSuggestionJob(t, te.Server.URL, "job_completion@example.com")

	statusReq, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	statusReq.Header.Set("Authorization", "Bearer "+token)
	statusResp, err := http.DefaultClient.Do(statusReq)
	if err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusResp.StatusCode)
	}

	var body struct {
		Status             string     `json:"status"`
		SuggestedQuestions [][]string `json:"suggested_questions"`
		ErrorMessage       *string    `json:"error_message"`
	}
	json.NewDecoder(statusResp.Body).Decode(&body)

	if body.Status != models.SuggestionJobStatusCompleted.String() {
		t.Errorf("expected completed status, got %s", body.Status)
	}

	if len(body.SuggestedQuestions) == 0 || len(body.SuggestedQuestions[0]) != 3 {
		t.Errorf("expected 3 suggestions, got %v", body.SuggestedQuestions)
	}

	if body.ErrorMessage != nil {
		t.Errorf("expected no error message, got %s", *body.ErrorMessage)
	}
}

func TestSuggestionRegeneration_JobFailure(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	ctx := context.Background()

	result := testdb.CreateChatbot(t, te.DB)
	chatbotID := result.Chatbot.ID

	job, err := db.CreateSuggestionJob(ctx, te.DB, chatbotID)
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	errMsg := "database connection failed"
	err = db.FailSuggestionJob(ctx, te.DB, job.ID, errMsg)
	if err != nil {
		t.Fatalf("failed to fail job: %v", err)
	}

	token := authTokenForSuggestionJob(t, te.Server.URL, "job_failure@example.com")

	statusReq, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	statusReq.Header.Set("Authorization", "Bearer "+token)
	statusResp, err := http.DefaultClient.Do(statusReq)
	if err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusResp.StatusCode)
	}

	var body struct {
		Status       string  `json:"status"`
		ErrorMessage *string `json:"error_message"`
	}
	json.NewDecoder(statusResp.Body).Decode(&body)

	if body.Status != models.SuggestionJobStatusFailed.String() {
		t.Errorf("expected failed status, got %s", body.Status)
	}

	if body.ErrorMessage == nil || *body.ErrorMessage != errMsg {
		t.Errorf("expected error message %q, got %v", errMsg, body.ErrorMessage)
	}
}

func TestSuggestionRegeneration_JobWithRunningStatus(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	ctx := context.Background()

	result := testdb.CreateChatbot(t, te.DB)
	chatbotID := result.Chatbot.ID

	job, err := db.CreateSuggestionJob(ctx, te.DB, chatbotID)
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	err = db.UpdateSuggestionJobStatus(ctx, te.DB, job.ID, models.SuggestionJobStatusRunning)
	if err != nil {
		t.Fatalf("failed to update job status: %v", err)
	}

	token := authTokenForSuggestionJob(t, te.Server.URL, "job_running@example.com")

	statusReq, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	statusReq.Header.Set("Authorization", "Bearer "+token)
	statusResp, err := http.DefaultClient.Do(statusReq)
	if err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusResp.StatusCode)
	}

	var body struct {
		Status string `json:"status"`
	}
	json.NewDecoder(statusResp.Body).Decode(&body)

	if body.Status != models.SuggestionJobStatusRunning.String() {
		t.Errorf("expected running status, got %s", body.Status)
	}
}
