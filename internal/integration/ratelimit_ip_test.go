package integration

import (
	"net/http"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/pkg/config"
)

func TestRateLimit_IPIsolationOnHealth(t *testing.T) {
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.RateLimitGlobalRequestsPerMinute = 2
		cfg.RateLimitGlobalWindowSeconds = 60
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// first two requests should be allowed
	for i := 0; i < 2; i++ {
		res, _ := testHTTPGet(te.Server.URL + "/health")
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		drainBody(res)
	}
	// third should be rate-limited
	res3, _ := testHTTPGet(te.Server.URL + "/health")
	if res3.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", res3.StatusCode)
	}
	res3.Body.Close()
}
