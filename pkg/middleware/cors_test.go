package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware_GET(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mw := CORSMiddleware("http://localhost:5173")(h)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("missing allow origin header")
	}
	if rr.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Fatalf("missing allow headers")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("missing allow methods")
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status = %d", rr.Result().StatusCode)
	}
}

func TestCORSMiddleware_OPTIONS(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mw := CORSMiddleware("http://localhost:5173")(h)
	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("missing allow origin header")
	}
	if rr.Result().StatusCode != http.StatusNoContent {
		t.Fatalf("preflight status = %d", rr.Result().StatusCode)
	}
}

func TestCORSMiddlewareAllowOrigins_AllowsMatchingOrigin(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mw := CORSMiddlewareAllowOrigins([]string{"http://localhost:5173", "https://app.example.com"})(h)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)
	if rr.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Fatalf("unexpected allow origin: %s", rr.Header().Get("Access-Control-Allow-Origin"))
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status = %d", rr.Result().StatusCode)
	}
}

func TestCORSMiddlewareAllowOrigins_DeniesUnknownOrigin(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mw := CORSMiddlewareAllowOrigins([]string{"http://localhost:5173"})(h)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "https://unknown.example")
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("should not set allow origin")
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status = %d", rr.Result().StatusCode)
	}
}
