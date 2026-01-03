package integration

import (
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestAuth_InvalidBearerFormat_Protected401(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Token abc.def")
	res, _ := testHTTPClient().Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	drainBody(res)
}

func TestAuth_BearerWithoutToken_Protected401(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer")
	res, _ := testHTTPClient().Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	drainBody(res)
}

func TestAuth_BasicScheme_Protected401(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Basic abc123")
	res, _ := testHTTPClient().Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	drainBody(res)
}
