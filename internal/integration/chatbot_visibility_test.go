package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
)

// TestChatbot_IsPublicToggle_ViaSecuritySettings covers the dashboard toggle:
// PUT /chatbots/{id}/security {"is_public": ...} hides and re-exposes the bot
// on the public widget endpoints.
func TestChatbot_IsPublicToggle_ViaSecuritySettings(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authTokenForHandoff(t, te.Server.URL, "visibility@example.com")
	cb, _ := json.Marshal(map[string]any{"name": "Visibility Bot", "language": "en-US"})
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, err := testHTTPClient().Do(reqC)
	if err != nil {
		t.Fatalf("create bot: %v", err)
	}
	defer drainBody(resC)
	var bot struct {
		ID       string `json:"id"`
		IsPublic bool   `json:"is_public"`
	}
	if err := json.NewDecoder(resC.Body).Decode(&bot); err != nil || bot.ID == "" {
		t.Fatalf("decode bot: %v", err)
	}
	if !bot.IsPublic {
		t.Fatal("new chatbots should be public by default")
	}

	publicStatus := func() int {
		res, err := testHTTPGet(te.Server.URL + "/api/v1/public/chatbots/" + bot.ID)
		if err != nil {
			t.Fatalf("public config: %v", err)
		}
		defer drainBody(res)
		return res.StatusCode
	}
	setPublic := func(v bool) bool {
		body := `{"is_public": false}`
		if v {
			body = `{"is_public": true}`
		}
		req, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/security", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		res, err := testHTTPClient().Do(req)
		if err != nil {
			t.Fatalf("update security: %v", err)
		}
		defer drainBody(res)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("update security failed: %d", res.StatusCode)
		}
		var updated struct {
			IsPublic bool `json:"is_public"`
		}
		if err := json.NewDecoder(res.Body).Decode(&updated); err != nil {
			t.Fatalf("decode update: %v", err)
		}
		return updated.IsPublic
	}

	if code := publicStatus(); code != http.StatusOK {
		t.Fatalf("public bot: expected 200, got %d", code)
	}
	if setPublic(false) {
		t.Fatal("expected is_public=false in update response")
	}
	if code := publicStatus(); code != http.StatusNotFound {
		t.Fatalf("private bot: expected 404, got %d", code)
	}
	if !setPublic(true) {
		t.Fatal("expected is_public=true in update response")
	}
	if code := publicStatus(); code != http.StatusOK {
		t.Fatalf("re-published bot: expected 200, got %d", code)
	}
}
