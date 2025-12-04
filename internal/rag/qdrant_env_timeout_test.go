package rag

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewQdrantClientFromEnv_TimeoutOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer srv.Close()
	t.Setenv("QDRANT_URL", srv.URL)
	t.Setenv("QDRANT_TIMEOUT_MS", "4200")
	c, err := NewQdrantClientFromEnv()
	if err != nil {
		t.Fatalf("client err: %v", err)
	}
	if c.http.Timeout.Milliseconds() != 4200 {
		t.Fatalf("timeout not applied")
	}
}
