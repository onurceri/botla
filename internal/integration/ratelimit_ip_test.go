package integration

import (
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestRateLimit_IPIsolationOnHealth(t *testing.T) {
	t.Setenv("RATE_LIMIT_GLOBAL_REQUESTS_PER_MINUTE", "2")
	t.Setenv("RATE_LIMIT_GLOBAL_WINDOW_SECONDS", "60")
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// first two requests should be allowed
	for i := 0; i < 2; i++ {
		res, _ := http.Get(te.Server.URL + "/health")
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		res.Body.Close()
	}
	// third should be rate-limited
	res3, _ := http.Get(te.Server.URL + "/health")
	if res3.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", res3.StatusCode)
	}
	res3.Body.Close()
}
