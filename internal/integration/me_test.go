package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestMe_ReturnsSubscriptionPlan(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "me@example.com")
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	defer res.Body.Close()

	var body struct {
		Usage struct {
			FilesCount    int `json:"files_count"`
			StorageUsedMB int `json:"storage_used_mb"`
			URLsCount     int `json:"urls_count"`
			TokensUsed    int `json:"tokens_used"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// New user should have 0 usage
	if body.Usage.FilesCount != 0 {
		t.Errorf("expected 0 files, got %d", body.Usage.FilesCount)
	}
	if body.Usage.URLsCount != 0 {
		t.Errorf("expected 0 urls, got %d", body.Usage.URLsCount)
	}
	if body.Usage.TokensUsed != 0 {
		t.Errorf("expected 0 tokens, got %d", body.Usage.TokensUsed)
	}
}
