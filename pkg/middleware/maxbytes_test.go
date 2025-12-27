package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMaxBytesMiddleware_AllowsSmallBody(t *testing.T) {
	mw := MaxBytesMiddleware(1024) // 1KB limit
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(body) != 100 {
			t.Errorf("Body length = %d, want 100", len(body))
		}
		w.WriteHeader(http.StatusOK)
	})

	body := strings.Repeat("x", 100) // 100 bytes
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rr.Code)
	}
}

func TestMaxBytesMiddleware_RejectsOversizedBody(t *testing.T) {
	mw := MaxBytesMiddleware(100) // 100 byte limit
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			// MaxBytesReader returns an error when limit is exceeded
			http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := strings.Repeat("x", 200) // 200 bytes, exceeds 100 limit
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Status = %d, want 413", rr.Code)
	}
}

func TestMaxBytesMiddleware_AllowsGetRequest(t *testing.T) {
	mw := MaxBytesMiddleware(100)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rr.Code)
	}
}

func TestMaxBytesMiddleware_PassesThrough(t *testing.T) {
	mw := MaxBytesMiddleware(1024)
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("small"))
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	if !called {
		t.Error("Handler was not called")
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("Status = %d, want 201", rr.Code)
	}
}
