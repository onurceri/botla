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

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrivacyFlow(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	require.NoError(t, err)
	defer fixtures.TeardownTestEnv(te)

	// Create admin user
	adminEmail := fmt.Sprintf("admin_%d@example.com", time.Now().UnixNano())
	adminID := registerUser(t, te.DB, te.Server.URL, adminEmail, "Test@123")

	// Make user an admin via DB
	_, err = te.DB.Exec("UPDATE users SET is_platform_admin = true WHERE id = $1", adminID)
	require.NoError(t, err)

	adminToken := loginUser(t, te.Server.URL, adminEmail, "Test@123")

	// Create regular user
	userEmail := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())
	userID := registerUser(t, te.DB, te.Server.URL, userEmail, "Test@123")
	userToken := loginUser(t, te.Server.URL, userEmail, "Test@123")
	var exportRequestID string

	t.Run("User Consents", func(t *testing.T) {
		// Update consents
		body := map[string]bool{
			"marketing":       true,
			"analytics":       false,
			"personalization": true,
			"third_party":     false,
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/me/privacy/consents", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Get consents
		req, _ = http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/privacy/consents", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err = testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var consents map[string]bool
		err = json.NewDecoder(resp.Body).Decode(&consents)
		require.NoError(t, err)

		assert.Equal(t, true, consents["marketing"])
		assert.Equal(t, false, consents["analytics"])
		assert.Equal(t, true, consents["personalization"])
		assert.Equal(t, false, consents["third_party"])
	})

	t.Run("User Request Export", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/export", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

		// It might return 200 OK or 201 Created
		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, resp.StatusCode)

		var pr repository.PrivacyRequest
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

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

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

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var pr repository.PrivacyRequest
		err = json.NewDecoder(resp.Body).Decode(&pr)
		require.NoError(t, err)
		assert.Equal(t, "correction", pr.RequestType)
		assert.Equal(t, "Correct my address to 123 Street", pr.Reason)
	})

	t.Run("Admin List Requests", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/admin/privacy/requests", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Data []repository.PrivacyRequest `json:"data"`
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

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

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

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.True(t, strings.Contains(string(b), `"email": "`+userEmail+`"`), "export should include user email")
	})

	t.Run("Admin Deny Request with Reason", func(t *testing.T) {
		// Fresh user: the shared user already has a pending export request, which
		// would make this POST fail with a 409 duplicate conflict.
		denyEmail := fmt.Sprintf("deny_%d@example.com", time.Now().UnixNano())
		registerUser(t, te.DB, te.Server.URL, denyEmail, "Test@123")
		denyToken := loginUser(t, te.Server.URL, denyEmail, "Test@123")

		// Create a new request to deny
		body := map[string]string{
			"reason": "Test reason",
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/export", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+denyToken)
		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		require.Contains(t, []int{http.StatusOK, http.StatusCreated}, resp.StatusCode)
		var pr repository.PrivacyRequest
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
		require.NotEmpty(t, pr.ID)
		drainBody(resp)

		// Deny it
		denyBody := map[string]string{
			"action":        "deny",
			"denial_reason": "Insufficient details provided",
		}
		dbDeny, err := json.Marshal(denyBody)
		require.NoError(t, err)
		req, _ = http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/admin/privacy/requests/"+pr.ID, bytes.NewReader(dbDeny))
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err = testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check status and denial reason in DB
		var status, denialReason string
		err = te.DB.QueryRow("SELECT status, denial_reason FROM privacy_requests WHERE id = $1", pr.ID).Scan(&status, &denialReason)
		require.NoError(t, err)
		assert.Equal(t, "denied", status)
		assert.Equal(t, "Insufficient details provided", denialReason)
	})

	t.Run("User List Privacy Requests", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/privacy/requests", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Data  []repository.PrivacyRequest `json:"data"`
			Total int                         `json:"total"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Should have at least the export request we created earlier
		found := false
		for _, r := range result.Data {
			if r.ID == exportRequestID {
				found = true
				break
			}
		}
		assert.True(t, found, "Export request should be in user's requests list")
		assert.GreaterOrEqual(t, result.Total, 1)
	})

	t.Run("Prevent Duplicate Export Requests", func(t *testing.T) {
		// Create a new user to test duplicate prevention
		dupUserEmail := fmt.Sprintf("dup_%d@example.com", time.Now().UnixNano())
		registerUser(t, te.DB, te.Server.URL, dupUserEmail, "Test@123")
		dupUserToken := loginUser(t, te.Server.URL, dupUserEmail, "Test@123")

		// First export request should succeed
		req1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/export", nil)
		req1.Header.Set("Authorization", "Bearer "+dupUserToken)

		resp1, err := testHTTPClient().Do(req1)
		require.NoError(t, err)
		defer drainBody(resp1)
		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, resp1.StatusCode)

		// Second export request should fail with 409 Conflict
		req2, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/export", nil)
		req2.Header.Set("Authorization", "Bearer "+dupUserToken)

		resp2, err := testHTTPClient().Do(req2)
		require.NoError(t, err)
		defer drainBody(resp2)
		assert.Equal(t, http.StatusConflict, resp2.StatusCode)
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

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

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

	t.Run("User List Privacy Requests with Pagination", func(t *testing.T) {
		// Fresh user: the shared user's account is deleted by "Admin Process Request".
		pagEmail := fmt.Sprintf("pag_%d@example.com", time.Now().UnixNano())
		pagUserID := registerUser(t, te.DB, te.Server.URL, pagEmail, "Test@123")
		pagToken := loginUser(t, te.Server.URL, pagEmail, "Test@123")

		// Seed >10 requests directly in the DB so the list spans two pages. We bypass
		// the API here because duplicate-prevention limits a user to one active request
		// per type, which could never reach the >10 rows needed to exercise pagination.
		seedTypes := []string{"correction", "deletion", "export"}
		for i := 0; i < 12; i++ {
			_, err := te.DB.Exec(
				`INSERT INTO privacy_requests (user_id, user_email, request_type, status, reason)
				 VALUES ($1, $2, $3, 'completed', $4)`,
				pagUserID, pagEmail, seedTypes[i%len(seedTypes)], fmt.Sprintf("seed %d", i),
			)
			require.NoError(t, err)
		}

		// Test first page with limit 10
		req1, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/privacy/requests?page=1&limit=10", nil)
		req1.Header.Set("Authorization", "Bearer "+pagToken)

		resp1, err := testHTTPClient().Do(req1)
		require.NoError(t, err)
		defer drainBody(resp1)
		assert.Equal(t, http.StatusOK, resp1.StatusCode)

		var result1 struct {
			Data  []repository.PrivacyRequest `json:"data"`
			Total int                         `json:"total"`
			Page  int                         `json:"page"`
			Limit int                         `json:"limit"`
		}
		err = json.NewDecoder(resp1.Body).Decode(&result1)
		require.NoError(t, err)

		// Some requests may be blocked by rate limits, just verify pagination works
		assert.GreaterOrEqual(t, len(result1.Data), 1)
		assert.Equal(t, 1, result1.Page)
		assert.Equal(t, 10, result1.Limit)
		assert.GreaterOrEqual(t, result1.Total, 10)

		// Test second page
		req2, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/privacy/requests?page=2&limit=10", nil)
		req2.Header.Set("Authorization", "Bearer "+pagToken)

		resp2, err := testHTTPClient().Do(req2)
		require.NoError(t, err)
		defer drainBody(resp2)
		assert.Equal(t, http.StatusOK, resp2.StatusCode)

		var result2 struct {
			Data  []repository.PrivacyRequest `json:"data"`
			Total int                         `json:"total"`
			Page  int                         `json:"page"`
			Limit int                         `json:"limit"`
		}
		err = json.NewDecoder(resp2.Body).Decode(&result2)
		require.NoError(t, err)

		assert.Equal(t, 2, result2.Page)
		assert.Equal(t, 10, result2.Limit)
		assert.Equal(t, result1.Total, result2.Total)
	})

	t.Run("User Delete Privacy Request", func(t *testing.T) {
		// Fresh user: the shared user's account is deleted by "Admin Process Request",
		// and it already has pending requests that would cause duplicate conflicts.
		delEmail := fmt.Sprintf("del_%d@example.com", time.Now().UnixNano())
		registerUser(t, te.DB, te.Server.URL, delEmail, "Test@123")
		delToken := loginUser(t, te.Server.URL, delEmail, "Test@123")

		// Create a new privacy request to delete
		body := map[string]string{
			"reason": "Test for deletion",
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/correction", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+delToken)

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var pr repository.PrivacyRequest
		err = json.NewDecoder(resp.Body).Decode(&pr)
		require.NoError(t, err)
		require.NotEmpty(t, pr.ID)

		// Delete the request
		deleteReq, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/me/privacy/requests/"+pr.ID, nil)
		deleteReq.Header.Set("Authorization", "Bearer "+delToken)

		deleteResp, err := testHTTPClient().Do(deleteReq)
		require.NoError(t, err)
		defer drainBody(deleteResp)
		assert.Equal(t, http.StatusOK, deleteResp.StatusCode)

		// Verify it's soft-deleted from DB (deleted_at is set)
		var count int
		err = te.DB.QueryRow("SELECT COUNT(*) FROM privacy_requests WHERE id = $1 AND deleted_at IS NULL", pr.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("User Cannot Delete Another User's Request", func(t *testing.T) {
		// Fresh actor user: the shared user's account is deleted by "Admin Process Request".
		actorEmail := fmt.Sprintf("actor_%d@example.com", time.Now().UnixNano())
		registerUser(t, te.DB, te.Server.URL, actorEmail, "Test@123")
		actorToken := loginUser(t, te.Server.URL, actorEmail, "Test@123")

		// Create another user
		otherUserEmail := fmt.Sprintf("other_%d@example.com", time.Now().UnixNano())
		registerUser(t, te.DB, te.Server.URL, otherUserEmail, "Test@123")
		otherUserToken := loginUser(t, te.Server.URL, otherUserEmail, "Test@123")

		// Create a privacy request for the other user
		body := map[string]string{
			"reason": "Other user's request",
		}
		b, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/me/privacy/correction", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+otherUserToken)

		resp, err := testHTTPClient().Do(req)
		require.NoError(t, err)
		defer drainBody(resp)

		var pr repository.PrivacyRequest
		err = json.NewDecoder(resp.Body).Decode(&pr)
		require.NoError(t, err)
		require.NotEmpty(t, pr.ID)

		// Try to delete with the actor's token (the request belongs to the other user)
		deleteReq, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/me/privacy/requests/"+pr.ID, nil)
		deleteReq.Header.Set("Authorization", "Bearer "+actorToken)

		deleteResp, err := testHTTPClient().Do(deleteReq)
		require.NoError(t, err)
		defer drainBody(deleteResp)
		assert.Equal(t, http.StatusNotFound, deleteResp.StatusCode)

		// Verify the request still exists
		var count int
		err = te.DB.QueryRow("SELECT COUNT(*) FROM privacy_requests WHERE id = $1", pr.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}
