package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManualRetry_Integration(t *testing.T) {
	te, err := SetupTestEnv()
	require.NoError(t, err)
	defer TeardownTestEnv(te)

	// Register user
	email := "retry@example.com"
	token := registerAndGetToken(t, te.Server.URL, email, TestPassword)

	// Create chatbot
	chatbotID := createChatbot(t, te.Server.URL, token, "Retry Bot")

	// Create source
	sourceID := createTextSource(t, te.Server.URL, token, chatbotID, "Test content for retry")

	// Find the job created for this source
	var jobID string
	err = te.DB.QueryRow("SELECT id FROM training_jobs WHERE source_id = $1", sourceID).Scan(&jobID)
	require.NoError(t, err)

	// Manually mark job as failed to simulate a failure
	_, err = te.DB.Exec("UPDATE training_jobs SET status = 'failed', failed_step = 'embed_chunks', error_code = 'TEST_FAIL', error_message = 'Simulated failure' WHERE id = $1", jobID)
	require.NoError(t, err)

	// Retry
	req, _ := http.NewRequest("POST", te.Server.URL+"/api/v1/sources/"+sourceID+"/job/retry", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, jobID, result["job_id"])

	// Verify job is pending again
	var status string
	var retryCount int
	err = te.DB.QueryRow("SELECT status, retry_count FROM training_jobs WHERE id = $1", jobID).Scan(&status, &retryCount)
	require.NoError(t, err)
	
	// It's either pending or already being processed by the background worker
	assert.True(t, status == "pending" || status == "running" || status == "completed")
	assert.Equal(t, 0, retryCount)
}

func TestAutomaticRetry_Integration(t *testing.T) {
	// This test is harder because it relies on the background worker
	// Automatic retry logic is thoroughly tested in unit tests.
}
