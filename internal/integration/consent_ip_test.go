package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsentIPExtraction(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	require.NoError(t, err)
	defer fixtures.TeardownTestEnv(te)

	// Create regular user
	userEmail := fmt.Sprintf("user_ip_%d@example.com", time.Now().UnixNano())
	userID := registerUser(t, te.DB, te.Server.URL, userEmail, "Test@123")
	userToken := loginUser(t, te.Server.URL, userEmail, "Test@123")

	t.Run("Extracts IP from X-Forwarded-For", func(t *testing.T) {
		marketing := true
		body := map[string]*bool{
			"marketing": &marketing,
		}
		b, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/me/privacy/consents", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+userToken)
		req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var ip string
		err = te.DB.QueryRow("SELECT ip_address FROM user_consents WHERE user_id = $1 AND consent_type = 'marketing'", userID).Scan(&ip)
		require.NoError(t, err)
		assert.Equal(t, "1.2.3.4", ip)
	})

	t.Run("Extracts IP from X-Real-IP", func(t *testing.T) {
		analytics := true
		body := map[string]*bool{
			"analytics": &analytics,
		}
		b, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/me/privacy/consents", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+userToken)
		req.Header.Set("X-Real-IP", "9.10.11.12")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var ip string
		err = te.DB.QueryRow("SELECT ip_address FROM user_consents WHERE user_id = $1 AND consent_type = 'analytics'", userID).Scan(&ip)
		require.NoError(t, err)
		assert.Equal(t, "9.10.11.12", ip)
	})

	t.Run("Handles IP:PORT in RemoteAddr", func(t *testing.T) {
		personalization := true
		body := map[string]*bool{
			"personalization": &personalization,
		}
		b, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/me/privacy/consents", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+userToken)
		// We can't easily change RemoteAddr of a real request in http.DefaultClient.Do
		// but the test server's RemoteAddr will be localhost:PORT.
		// If it works without error, it means net.SplitHostPort worked.

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var ip string
		err = te.DB.QueryRow("SELECT ip_address FROM user_consents WHERE user_id = $1 AND consent_type = 'personalization'", userID).Scan(&ip)
		require.NoError(t, err)
		assert.Contains(t, []string{"127.0.0.1", "::1"}, ip)
	})

	t.Run("Returns 500 on invalid IP (DB error)", func(t *testing.T) {
		marketing := true
		body := map[string]*bool{
			"marketing": &marketing,
		}
		b, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/me/privacy/consents", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+userToken)
		req.Header.Set("X-Forwarded-For", "not-an-ip")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
