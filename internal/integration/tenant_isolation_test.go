package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestMultiTenantIsolation(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create User 1 and their chatbot
	token1 := authToken(t, te.Server.URL, "user1@example.com")
	create1 := map[string]any{"name": "User 1 Bot"}
	cbj1, _ := json.Marshal(create1)
	reqC1, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj1))
	reqC1.Header.Set("Authorization", "Bearer "+token1)
	reqC1.Header.Set("Content-Type", "application/json")
	resC1, _ := testHTTPClient().Do(reqC1)
	var bot1 chatbot
	json.NewDecoder(resC1.Body).Decode(&bot1)
	resC1.Body.Close()

	// Create User 2
	token2 := authToken(t, te.Server.URL, "user2@example.com")

	t.Run("User 2 cannot access User 1 chatbot", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot1.ID, nil)
		req.Header.Set("Authorization", "Bearer "+token2)
		res, err := testHTTPClient().Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
		drainBody(res)
	})

	t.Run("User 2 cannot update User 1 chatbot", func(t *testing.T) {
		update := map[string]any{"name": "Hacked Bot"}
		uj, _ := json.Marshal(update)
		req, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot1.ID, bytes.NewReader(uj))
		req.Header.Set("Authorization", "Bearer "+token2)
		req.Header.Set("Content-Type", "application/json")
		res, err := testHTTPClient().Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
		drainBody(res)
	})

	t.Run("User 2 cannot delete User 1 chatbot", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/chatbots/"+bot1.ID, nil)
		req.Header.Set("Authorization", "Bearer "+token2)
		res, err := testHTTPClient().Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
		drainBody(res)
	})
}
