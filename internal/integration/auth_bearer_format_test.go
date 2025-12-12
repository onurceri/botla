package integration

import (
	"net/http"
	"testing"
)

func TestAuth_InvalidBearerFormat_Protected401(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Token abc.def")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestAuth_BearerWithoutToken_Protected401(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestAuth_BasicScheme_Protected401(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Basic abc123")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}
