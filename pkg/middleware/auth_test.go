package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/auth"
)

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	mw := AuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidTokenType(t *testing.T) {
	mw := AuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	tok, _ := auth.GenerateToken("secret", "u1", "refresh", time.Minute)
	req.Header.Set("Authorization", "Bearer "+tok)
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ValidAccess(t *testing.T) {
	mw := AuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	tok, _ := auth.GenerateToken("secret", "u1", "access", time.Minute)
	req.Header.Set("Authorization", "Bearer "+tok)
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestOptionalAuthMiddleware_MissingHeader(t *testing.T) {
	mw := OptionalAuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserIDFromContext(r.Context()); ok {
			t.Error("expected no user ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestOptionalAuthMiddleware_ValidAccess(t *testing.T) {
	mw := OptionalAuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	tok, _ := auth.GenerateToken("secret", "u1", "access", time.Minute)
	req.Header.Set("Authorization", "Bearer "+tok)
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := UserIDFromContext(r.Context())
		if !ok || uid != "u1" {
			t.Errorf("expected user ID u1, got %v (ok=%v)", uid, ok)
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	mw := OptionalAuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserIDFromContext(r.Context()); ok {
			t.Error("expected no user ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
