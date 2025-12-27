package integration

import (
	"testing"
	"time"
)

func TestWorkerPool_ParallelProcessing(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Create user and get token
	token := registerAndGetToken(t, te.Server.URL, "workertest@example.com", TestPassword)

	// Create chatbot
	chatbotID := createChatbot(t, te.Server.URL, token, "Worker Test Bot")

	// Create 5 text sources concurrently
	sourceCount := 5
	sourceChan := make(chan string, sourceCount)

	for i := 0; i < sourceCount; i++ {
		go func(idx int) {
			content := "Parallel processing test content"
			id := createTextSource(t, te.Server.URL, token, chatbotID, content)
			sourceChan <- id
		}(i)
	}

	// Collected source IDs
	var sourceIDs []string
	for i := 0; i < sourceCount; i++ {
		sourceIDs = append(sourceIDs, <-sourceChan)
	}

	// Verify all jobs eventually complete
	// We wait up to 5 seconds
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for jobs to complete")
		case <-ticker.C:
			completed := 0
			for _, id := range sourceIDs {
				// Check source status in DB directly or via API
				// Using DB for speed
				// We need to find the job associated with this source
				// Since we don't have job ID directly from createTextSource (it returns source ID),
				// we query the job table.
				
				// Assuming one job per source for this test
				// Actually, job tracking is separate. We can check source processing status as proxy, 
				// or fetch the job.
				
				// Let's check source processing status from data_sources table
				var processingStatus string
				err := te.DB.QueryRow("SELECT processing_status FROM data_sources WHERE id=$1", id).Scan(&processingStatus)
				if err != nil {
					t.Fatalf("failed to query source status: %v", err)
				}
				if processingStatus == "completed" || processingStatus == "failed" {
					completed++
				}
			}
			if completed == sourceCount {
				return // Success
			}
		}
	}
}
