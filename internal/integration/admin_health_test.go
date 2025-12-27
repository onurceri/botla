package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminHealth(t *testing.T) {
	te, err := SetupTestEnv()
	require.NoError(t, err)
	defer TeardownTestEnv(te)

	// Create admin user
	adminEmail := fmt.Sprintf("admin_%d@example.com", time.Now().UnixNano())
	adminID := registerUser(t, te.DB, te.Server.URL, adminEmail, "Test@123")

	// Make user an admin via DB
	_, err = te.DB.Exec("UPDATE users SET is_platform_admin = true WHERE id = $1", adminID)
	require.NoError(t, err)

	adminToken := loginUser(t, te.Server.URL, adminEmail, "Test@123")

	// Create regular user
	userEmail := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())
	registerUser(t, te.DB, te.Server.URL, userEmail, "Test@123")
	userToken := loginUser(t, te.Server.URL, userEmail, "Test@123")

	t.Run("Get Detailed Health - Admin", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/health/detailed", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]any
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)

		// Basic structure check
		assert.Contains(t, health, "status")
		assert.Contains(t, health, "version")
		assert.Contains(t, health, "uptime")
		assert.Contains(t, health, "environment")
		assert.Contains(t, health, "dependencies")

		// Check dependencies
		deps, ok := health["dependencies"].([]any)
		assert.True(t, ok)
		assert.NotEmpty(t, deps)

		// Verify Postgres is present and OK (since we have a DB)
		foundPostgres := false
		for _, d := range deps {
			dep := d.(map[string]any)
			if dep["name"] == "postgres" {
				foundPostgres = true
				assert.Equal(t, "ok", dep["status"])
				break
			}
		}
		assert.True(t, foundPostgres, "Postgres dependency should be reported")
	})

	t.Run("Get Detailed Health - Regular User Forbidden", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/health/detailed", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
