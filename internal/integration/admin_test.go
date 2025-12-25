package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminFlow(t *testing.T) {
	te, err := SetupTestEnv()
	require.NoError(t, err)
	defer TeardownTestEnv(te)

	// 1. Create Admin User
	adminEmail := fmt.Sprintf("admin_%d@example.com", time.Now().UnixNano())
	adminID := registerUser(t, te.DB, te.Server.URL, adminEmail, "password123")

	// Make user an admin via DB
	_, err = te.DB.Exec("UPDATE users SET is_platform_admin = true WHERE id = $1", adminID)
	require.NoError(t, err)

	adminToken := loginUser(t, te.Server.URL, adminEmail, "password123")

	// 2. Create Regular User
	userEmail := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())
	userID := registerUser(t, te.DB, te.Server.URL, userEmail, "password123")
	userToken := loginUser(t, te.Server.URL, userEmail, "password123")

	t.Run("Get Overview Stats", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/stats/overview", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats map[string]any
		err = json.NewDecoder(resp.Body).Decode(&stats)
		require.NoError(t, err)

		// Check keys exist
		assert.Contains(t, stats, "total_users")
		assert.Contains(t, stats, "total_organizations")
		assert.Contains(t, stats, "total_chatbots")
		assert.Contains(t, stats, "total_messages")
	})

	t.Run("List Users", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Users []models.User `json:"users"`
			Total int           `json:"total"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, result.Total, 2) // At least admin + user

		foundAdmin := false
		foundUser := false
		for _, u := range result.Users {
			if u.ID == adminID {
				foundAdmin = true
			}
			if u.ID == userID {
				foundUser = true
			}
		}
		assert.True(t, foundAdmin, "Admin user should be in the list")
		assert.True(t, foundUser, "Regular user should be in the list")
	})

	t.Run("Get Specific User", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/users/"+userID, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var user models.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, userEmail, user.Email)
	})

	t.Run("List Organizations", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/organizations", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Unauthorized Access (Regular User)", func(t *testing.T) {
		endpoints := []string{
			"/api/v1/admin/stats/overview",
			"/api/v1/admin/users",
			"/api/v1/admin/organizations",
		}

		for _, endpoint := range endpoints {
			req, _ := http.NewRequest(http.MethodGet, te.Server.URL+endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+userToken)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			resp.Body.Close()

			// Should be Forbidden (403) or Unauthorized (401) depending on implementation
			// RequirePlatformAdmin usually returns 403 for authenticated non-admins
			assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Endpoint: "+endpoint)
		}
	})

	t.Run("Unauthorized Access (No Token)", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/stats/overview", nil)
		// No auth header

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func registerUser(t *testing.T, db *sql.DB, baseURL, email, password string) string {
	body := map[string]string{
		"email":     email,
		"password":  password,
		"full_name": "Test User",
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]any
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("Failed to register user: %d %v", resp.StatusCode, errResp)
	}

	var id string
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&id)
	require.NoError(t, err)
	return id
}

func loginUser(t *testing.T, baseURL, email, password string) string {
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to login user: %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result.Token
}
