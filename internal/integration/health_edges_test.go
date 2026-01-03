package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
)

func TestHealth_QdrantDown_503(t *testing.T) {
	t.Parallel()
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	defer bad.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.QDRANT_URL = bad.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	res, _ := testHTTPGet(te.Server.URL + "/health")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.StatusCode)
	}
	drainBody(res)
}

func TestHealth_DBDown_503(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	te.DB.Close()
	res, _ := testHTTPGet(te.Server.URL + "/health")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.StatusCode)
	}
	drainBody(res)
}
