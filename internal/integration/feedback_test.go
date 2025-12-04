package integration

import (
	"net/http"
	"testing"
)

func TestFeedback_Protected_Unauthorized(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/messages/unknown/feedback", nil)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
}

func TestFeedback_Protected_Authorized(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	token := authToken(t, te.Server.URL, "fb@example.com")
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/messages/unknown/feedback", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode == http.StatusUnauthorized {
		t.Fatalf("expected not 401")
	}
}
