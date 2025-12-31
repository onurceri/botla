package rag

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewQdrantClient_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer srv.Close()

	c, err := NewQdrantClient(&QdrantConfig{URL: srv.URL, Timeout: 4200 * time.Millisecond})
	if err != nil {
		t.Fatalf("client err: %v", err)
	}
	if c.http.Timeout.Milliseconds() != 4200 {
		t.Fatalf("timeout not applied")
	}
}
