package integration

import (
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestHealth(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	res, err := testHTTPGet(te.Server.URL + "/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	drainBody(res)
}
