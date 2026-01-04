package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func TestRecovery_Prod(t *testing.T) {
	log := logger.New("ERROR")
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("oops")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.RecoveryMiddleware(log, "production")(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	// Check JSON response
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("expected json response")
	}
	if resp["code"] != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR code")
	}
	if strings.Contains(rec.Body.String(), "oops") {
		t.Fatalf("stack trace leaked in prod")
	}
}

func TestRecovery_Dev(t *testing.T) {
	log := logger.New("ERROR")
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("oops")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.RecoveryMiddleware(log, "development")(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	// Check Stack Trace
	body := rec.Body.String()
	if !strings.Contains(body, "Panic recovered: oops") {
		t.Fatalf("expected panic message in dev")
	}
	if !strings.Contains(body, "recovery_test.go") {
		t.Fatalf("expected stack trace in dev")
	}
}
