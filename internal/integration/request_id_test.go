package integration

import (
	"net/http"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
)

func TestRequestID_Integration(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Test 1: Response includes generated request ID
	resp, err := testHTTPGet(te.Server.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer drainBody(resp)

	reqID := resp.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Error("expected X-Request-ID header in response")
	}

	// Verify it looks like a UUID (36 characters)
	if len(reqID) != 36 {
		t.Errorf("expected UUID format (36 chars), got %s (%d chars)", reqID, len(reqID))
	}

	// Test 2: Provided request ID is returned
	client := &http.Client{}
	req, _ := http.NewRequest("GET", te.Server.URL+"/health", nil)
	req.Header.Set("X-Request-ID", "test-id-12345")

	resp2, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.Header.Get("X-Request-ID") != "test-id-12345" {
		t.Error("expected provided request ID to be returned")
	}
}
