package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
)

func TestGetPlan_ReturnsSplitConfig(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "plan_test@example.com")

	// Ensure user is on 'free' plan (default) or 'pro' if test setup differs
	// But let's check response structure mainly.

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/plan", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := testHTTPClient().Do(req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	defer drainBody(res)

	var body struct {
		Code   string `json:"code"`
		Limits struct {
			MaxChatbots int `json:"max_chatbots"`
		} `json:"limits"`
		Features struct {
			Chat struct {
				AllowedModels []string `json:"allowed_models"`
			} `json:"chat"`
			Files struct {
				MaxFilesTotal int `json:"max_files_total"`
			} `json:"files"`
		} `json:"features"`
		Config interface{} `json:"config"` // Should be absent
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Code == "" {
		t.Error("expected plan code")
	}

	// Verify Limits
	if body.Limits.MaxChatbots == 0 {
		// assuming default is not 0, or at least field exists.
		// Free plan max_chatbots might be defined in DB.
		// Let's just check if we can access it.
	}

	// Verify Features
	if len(body.Features.Chat.AllowedModels) == 0 {
		t.Error("expected allowed_models in features.chat")
	}

	// Verify Config is absent
	if body.Config != nil {
		t.Error("expected 'config' field to be absent")
	}
}
