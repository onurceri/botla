package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminQueueHandlers(t *testing.T) {
	pool := testdb.OpenParallelTestDB(t)

	log := logger.New("INFO")
	adminRepo := repository.NewPostgresAdminRepo(pool)
	adminSvc := services.NewAdminService(adminRepo, log)
	queueRepo := repository.NewPostgresQueueRepo(pool)
	sourceRepo := repository.NewPostgresSourceRepo(pool)
	h := NewAdminQueueHandlers(adminSvc, queueRepo, sourceRepo)
	ctx := context.Background()

	// Seed required data for foreign keys
	langID := uuid.New().String()
	_, err := pool.ExecContext(ctx, "INSERT INTO languages (id, code, name) VALUES ($1, 'en-test', 'English Test')", langID)
	require.NoError(t, err)

	planID := uuid.New().String()
	_, err = pool.ExecContext(ctx, "INSERT INTO plans (id, code, price) VALUES ($1, 'free-test', 0)", planID)
	require.NoError(t, err)

	userID := uuid.New().String()
	_, err = pool.ExecContext(ctx, "INSERT INTO users (id, email, password_hash, plan_id) VALUES ($1, 'test-queues@example.com', 'hash', $2)", userID, planID)
	require.NoError(t, err)

	chatbotID := uuid.New().String()
	_, err = pool.ExecContext(ctx, "INSERT INTO chatbots (id, user_id, name) VALUES ($1, $2, 'Test Bot')", chatbotID, userID)
	require.NoError(t, err)

	// Seed some data
	// 1. A pending source
	_, err = pool.ExecContext(ctx, `
		INSERT INTO data_sources (id, chatbot_id, source_type, status, created_at)
		VALUES (gen_random_uuid(), $1, 'url', 'pending', NOW())
	`, chatbotID)
	require.NoError(t, err)

	// 2. A processing source (stuck)
	stuckID := uuid.New().String()
	_, err = pool.ExecContext(ctx, `
		INSERT INTO data_sources (id, chatbot_id, source_type, status, created_at, last_refreshed_at)
		VALUES ($1, $2, 'url', 'processing', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour')
	`, stuckID, chatbotID)
	require.NoError(t, err)

	t.Run("GetQueues", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/queues", nil)
		rr := httptest.NewRecorder()

		h.GetQueues(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var stats []repository.QueueStats
		err := json.Unmarshal(rr.Body.Bytes(), &stats)
		require.NoError(t, err)

		found := false
		for _, s := range stats {
			if s.QueueName == "source_processing" {
				found = true
				assert.Equal(t, 1, s.PendingCount)
				assert.Equal(t, 1, s.ProcessingCount)
			}
		}
		assert.True(t, found)
	})

	t.Run("GetStuckJobs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/queues/stuck?threshold=30m", nil)
		rr := httptest.NewRecorder()

		h.GetStuckJobs(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var jobs []repository.StuckJob
		err := json.Unmarshal(rr.Body.Bytes(), &jobs)
		require.NoError(t, err)

		assert.Len(t, jobs, 1)
		assert.Equal(t, stuckID, jobs[0].ID)
	})

	t.Run("RetryJob", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/admin/queues/"+stuckID+"/retry", nil)
		// Mux stores path values in the request context
		req.SetPathValue("id", stuckID)
		rr := httptest.NewRecorder()

		h.RetryJob(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify status in DB
		var status string
		err := pool.QueryRowContext(ctx, "SELECT status FROM data_sources WHERE id = $1", stuckID).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, "pending", status)
	})
}
