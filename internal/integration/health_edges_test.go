package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestHealth_QdrantDown_503(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	t.Setenv("QDRANT_URL", bad.URL)
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer bad.Close()
	res, _ := http.Get(te.Server.URL + "/health")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestHealth_DBDown_503(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	te.DB.Close()
	res, _ := http.Get(te.Server.URL + "/health")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.StatusCode)
	}
	res.Body.Close()
}
