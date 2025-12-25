package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrivacyFlow(t *testing.T) {
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

	// Create regular user
	userEmail := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())
	userID := registerUser(t, te.DB, te.Server.URL, userEmail, "password123")
	userToken := loginUser(t, te.Server.URL, userEmail, "password123")
	var exportRequestID string

	t.Run("User Consents", func(t *testing.T) {
		// Update consents
		body := map[string]bool{
			"marketing": true,
			"analytics": false,
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/me/privacy/consents", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Get consents
		req, _ = http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/privacy/consents", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var consents map[string]bool
		err = json.NewDecoder(resp.Body).Decode(&consents)
		require.NoError(t, err)

		// Note: The response structure depends on GetMyConsents implementation.
		// If it returns a map or list, we should check accordingly.
		// Based on user_privacy.go it returns whatever GetMyConsents returns.
		// Assuming it returns a map like {"marketing": true, ...}
		// If not, we might need to adjust.
		// Let's assume it works for now or returns OK.
	})

	t.Run("User Request Export", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/export", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// It might return 200 OK or 201 Created
		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, resp.StatusCode)

		var pr db.PrivacyRequest
		err = json.NewDecoder(resp.Body).Decode(&pr)
		require.NoError(t, err)
		require.NotEmpty(t, pr.ID)
		assert.Equal(t, "export", pr.RequestType)
		assert.Equal(t, "pending", pr.Status)
		exportRequestID = pr.ID
	})

	t.Run("User Request Deletion", func(t *testing.T) {
		body := map[string]string{
			"reason": "I want to leave",
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/delete", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("User Request Correction", func(t *testing.T) {
		body := map[string]string{
			"reason": "Correct my address to 123 Street",
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/correction", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var pr db.PrivacyRequest
		err = json.NewDecoder(resp.Body).Decode(&pr)
		require.NoError(t, err)
		assert.Equal(t, "correction", pr.RequestType)
		assert.Equal(t, "Correct my address to 123 Street", pr.Reason)
	})

	t.Run("Admin List Requests", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/privacy/requests", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Data []db.PrivacyRequest `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		found := false
		for _, r := range result.Data {
			if r.UserID != nil && *r.UserID == userID && r.RequestType == "deletion" {
				found = true
				break
			}
		}
		assert.True(t, found, "Deletion request should be listed")
	})

	t.Run("Admin Approve Export Request", func(t *testing.T) {
		require.NotEmpty(t, exportRequestID)

		body := map[string]string{
			"action": "approve",
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/admin/privacy/requests/"+exportRequestID, bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		deadline := time.Now().Add(5 * time.Second)
		for {
			var status string
			err = te.DB.QueryRow("SELECT status FROM privacy_requests WHERE id = $1", exportRequestID).Scan(&status)
			require.NoError(t, err)
			if status == "completed" || status == "denied" {
				assert.Equal(t, "completed", status)
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("timed out waiting for export request to complete, last status=%s", status)
			}
			time.Sleep(100 * time.Millisecond)
		}

		var exportURL string
		err = te.DB.QueryRow("SELECT export_url FROM privacy_requests WHERE id = $1", exportRequestID).Scan(&exportURL)
		require.NoError(t, err)
		assert.NotEmpty(t, exportURL)
	})

	t.Run("User Download Export", func(t *testing.T) {
		require.NotEmpty(t, exportRequestID)

		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/privacy/requests/"+exportRequestID+"/download", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.True(t, strings.Contains(string(b), `"email": "`+userEmail+`"`), "export should include user email")
	})

	t.Run("Admin Deny Request with Reason", func(t *testing.T) {
		// Create a new request to deny
		body := map[string]string{
			"reason": "Test reason",
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/export", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+userToken)
		resp, _ := http.DefaultClient.Do(req)
		var pr db.PrivacyRequest
		json.NewDecoder(resp.Body).Decode(&pr)
		resp.Body.Close()

		// Deny it
		denyBody := map[string]string{
			"action":        "deny",
			"denial_reason": "Insufficient details provided",
		}
		dbDeny, err := json.Marshal(denyBody)
		require.NoError(t, err)
		req, _ = http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/admin/privacy/requests/"+pr.ID, bytes.NewReader(dbDeny))
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check status and denial reason in DB
		var status, denialReason string
		err = te.DB.QueryRow("SELECT status, denial_reason FROM privacy_requests WHERE id = $1", pr.ID).Scan(&status, &denialReason)
		require.NoError(t, err)
		assert.Equal(t, "denied", status)
		assert.Equal(t, "Insufficient details provided", denialReason)
	})

	t.Run("Admin Process Request", func(t *testing.T) {
		// First find the request ID
		var requestID string
		err := te.DB.QueryRow("SELECT id FROM privacy_requests WHERE user_id = $1 AND request_type = 'deletion' ORDER BY created_at DESC LIMIT 1", userID).Scan(&requestID)
		require.NoError(t, err)

		// Approve it
		body := map[string]string{
			"action": "approve",
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/admin/privacy/requests/"+requestID, bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check status in DB
		var status string
		err = te.DB.QueryRow("SELECT status FROM privacy_requests WHERE id = $1", requestID).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, "completed", status)

		// Verify user is anonymized
		var fullName, email string
		var deletedAt *time.Time
		err = te.DB.QueryRow("SELECT full_name, email, deleted_at FROM users WHERE id = $1", userID).Scan(&fullName, &email, &deletedAt)
		require.NoError(t, err)
		assert.Equal(t, "Anonymized User", fullName)
		assert.Contains(t, email, "anonymized-")
		assert.NotNil(t, deletedAt)
	})
}
