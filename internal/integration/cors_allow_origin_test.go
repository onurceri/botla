package integration

import (
	"net/http"
	"testing"
)

func TestCORS_AllowConfiguredOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if res.Header.Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("missing allow origin header")
	}
	res.Body.Close()
}
