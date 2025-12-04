package integration

import (
	"net/http"
	"testing"
)

func TestAnalytics_Smoke(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	token := authToken(t, te.Server.URL, "an@example.com")
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/analytics", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}
