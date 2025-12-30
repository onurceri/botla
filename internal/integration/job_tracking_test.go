package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestJobTracking_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	te, err := fixtures.SetupTestEnvWithMocks()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "jobtrack@example.com")
	chatbotID := createChatbot(t, te.Server.URL, token, "Job Track Bot")

	// Create source
	// Create source
	resp := createURLSource(t, te.Server.URL, token, chatbotID, "https://example.com")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create source: %d", resp.StatusCode)
	}
	var src struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&src); err != nil {
		t.Fatalf("failed to decode source response: %v", err)
	}
	resp.Body.Close()
	sourceID := src.ID

	// Poll for job status
	var lastStatus string
	// var lastStep string (unused)
	var attempts int

	// Wait up to 10 seconds (processing might take a moment in integration tests)
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-timeout:
			t.Errorf("job did not complete in time, last status: %s", lastStatus)
			break LOOP
		case <-ticker.C:
			req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceID+"/job", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				continue
			}

			if resp.StatusCode == http.StatusNotFound {
				// Job might not be created instantly if using async queue, though EnqueueSource creates it before returning
				// But maybe the API endpoint queries it?
				resp.Body.Close()
				continue
			}

			if resp.StatusCode != http.StatusOK {
				// retry
				resp.Body.Close()
				attempts++
				if attempts > 50 {
					t.Fatalf("failed to get job status, status code: %d", resp.StatusCode)
				}
				continue
			}

			var job map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			if status, ok := job["status"].(string); ok {
				lastStatus = status
				// if step, ok := job["current_step"].(string); ok {
				// 	lastStep = step
				// }

				if lastStatus == "completed" || lastStatus == "failed" {
					break LOOP
				}
			}
		}
	}

	if lastStatus != "completed" {
		t.Errorf("expected job to complete, got %s", lastStatus)
	}

	// Verify we can see the completed steps
	if lastStatus == "completed" {
		// Could verify other fields like started_at, completed_at if the API exposes them
	}
}
