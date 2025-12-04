package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/pkg/logger"
)

func TestRequestLogger_DefaultStatusOnWrite(t *testing.T) {
	log := logger.New("DEBUG")
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	RequestLogger(log)(h).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestRequestLogger_ExplicitStatus(t *testing.T) {
	log := logger.New("INFO")
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/y", nil)
	RequestLogger(log)(h).ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("want 418, got %d", rec.Code)
	}
}
