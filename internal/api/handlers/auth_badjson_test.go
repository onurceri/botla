package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterHandler_BadJSON(t *testing.T) {
	h := &AuthHandlers{Secret: "s"}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader("{"))
	h.RegisterHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
}

func TestLoginHandler_BadJSON(t *testing.T) {
	h := &AuthHandlers{Secret: "s"}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("{"))
	h.LoginHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
}

func TestRefreshHandler_BadJSON(t *testing.T) {
	h := &AuthHandlers{Secret: "s"}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader("{"))
	h.RefreshHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
}

func TestLogoutHandler_BadJSON(t *testing.T) {
	h := &AuthHandlers{Secret: "s"}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader("{"))
	h.LogoutHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
}

func TestRegisterHandler_MethodNotAllowed(t *testing.T) {
	h := &AuthHandlers{Secret: "s"}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/register", nil)
	h.RegisterHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rr.Code)
	}
}
