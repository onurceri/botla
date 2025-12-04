package integration

import (
	"net/http"
	"testing"
)

func TestHealth(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	res, err := http.Get(te.Server.URL + "/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	res.Body.Close()
}
