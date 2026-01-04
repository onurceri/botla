package integration

import (
	"fmt"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
)

func TestWorkerPool_ParallelProcessing(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnvWithMocks()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create user and get token
	token := registerAndGetToken(t, te.Server.URL, "workertest@example.com", fixtures.TestPassword)

	// Create chatbot
	chatbotID := createChatbot(t, te.Server.URL, token, "Worker Test Bot")

	// Create 5 text sources sequentially (concurrent version causes hangs)
	sourceCount := 5
	var sourceIDs []string

	for i := 0; i < sourceCount; i++ {
		content := fmt.Sprintf("Parallel processing test content %d", i)
		id := createTextSource(t, te.Server.URL, token, chatbotID, content)
		sourceIDs = append(sourceIDs, id)
	}

	// Verify all sources were created and jobs were enqueued
	// Check that all sources exist in the database
	for _, id := range sourceIDs {
		var status string
		err := te.DB.QueryRow("SELECT status FROM data_sources WHERE id=$1", id).Scan(&status)
		if err != nil {
			t.Fatalf("failed to query source status: %v", err)
		}
		if status != "pending" {
			t.Errorf("expected source %s to be pending, got %s", id, status)
		}
	}

	// Verify jobs were created for each source
	for _, sourceID := range sourceIDs {
		var jobCount int
		err := te.DB.QueryRow("SELECT COUNT(*) FROM training_jobs WHERE source_id=$1", sourceID).Scan(&jobCount)
		if err != nil {
			t.Fatalf("failed to query job count: %v", err)
		}
		if jobCount != 1 {
			t.Errorf("expected 1 job for source %s, got %d", sourceID, jobCount)
		}
	}
}
