package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminQueues(t *testing.T) {
	te, err := SetupTestEnv()
	require.NoError(t, err)
	defer TeardownTestEnv(te)

	// Create admin user
	adminEmail := fmt.Sprintf("admin_%d@example.com", time.Now().UnixNano())
	adminID := registerUser(t, te.DB, te.Server.URL, adminEmail, "password123")

	// Make user an admin via DB
	_, err = te.DB.Exec("UPDATE users SET is_platform_admin = true WHERE id = $1", adminID)
	require.NoError(t, err)

	adminToken := loginUser(t, te.Server.URL, adminEmail, "password123")

	// Seed required data for foreign keys
	chatbotID := "00000000-0000-0000-0000-00000000000a" // Use a unique ID to avoid conflict with dummy data
	// We need a valid user for chatbot
	_, err = te.DB.Exec("INSERT INTO chatbots (id, user_id, name) VALUES ($1, $2, 'Test Bot')", chatbotID, adminID)
	require.NoError(t, err)

	// Seed queue data
	// 1. A pending source
	_, err = te.DB.Exec(`
		INSERT INTO data_sources (id, chatbot_id, source_type, status, created_at)
		VALUES (gen_random_uuid(), $1, 'url', 'pending', NOW())
	`, chatbotID)
	require.NoError(t, err)

	// 2. A processing source (stuck)
	stuckID := "00000000-0000-0000-0000-000000000002"
	_, err = te.DB.Exec(`
		INSERT INTO data_sources (id, chatbot_id, source_type, status, created_at, last_refreshed_at)
		VALUES ($1, $2, 'url', 'processing', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour')
	`, stuckID, chatbotID)
	require.NoError(t, err)

	t.Run("GetQueues", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/queues", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats []db.QueueStats
		err = json.NewDecoder(resp.Body).Decode(&stats)
		require.NoError(t, err)

		found := false
		for _, s := range stats {
			if s.QueueName == "source_processing" {
				found = true
				// We expect at least 1 pending and 1 processing (from our seed)
				// Note: previous tests might have left data if not fully isolated, but TeardownTestEnv drops schema?
				// No, SetupTestEnv uses a transaction or clean DB.
				// But admin_test.go runs against the same DB potentially if parallel?
				// SetupTestEnv creates a NEW schema for each test usually or truncates.
				// Let's assume isolation.
				assert.GreaterOrEqual(t, s.PendingCount, 1)
				assert.GreaterOrEqual(t, s.ProcessingCount, 1)
			}
		}
		assert.True(t, found, "source_processing queue should be present")
	})

	t.Run("GetStuckJobs", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/queues/stuck?threshold=30m", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var jobs []db.StuckJob
		err = json.NewDecoder(resp.Body).Decode(&jobs)
		require.NoError(t, err)

		// Find our stuck job
		found := false
		for _, j := range jobs {
			if j.ID == stuckID {
				found = true
				break
			}
		}
		assert.True(t, found, "Stuck job should be returned")
	})

	t.Run("RetryJob", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/admin/queues/"+stuckID+"/retry", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify status in DB
		var status string
		err = te.DB.QueryRow("SELECT status FROM data_sources WHERE id = $1", stuckID).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, "pending", status)
	})
}
