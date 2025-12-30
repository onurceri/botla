package integration

import (
	"net/http"
	"strings"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestCORS_PatchAllowed(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodOptions, te.Server.URL+"/api/v1/organizations/123/members/456", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "PATCH")

	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}

	allowMethods := res.Header.Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Fatalf("missing Access-Control-Allow-Methods header")
	}

	// Check if PATCH is in the allowed methods
	// The header value is usually a comma-separated list
	if !contains(allowMethods, "PATCH") {
		t.Errorf("PATCH not found in Access-Control-Allow-Methods: %s", allowMethods)
	}
}

func contains(s, substr string) bool {
	// Simple check, in real implementation we might want to split by comma and trim
	// But since we know the exact string we put in, this is fine for now,
	// or we can use strings.Contains
	return strings.Contains(s, substr)
}
