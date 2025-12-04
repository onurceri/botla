package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProtectedHandler_NoContext(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	ProtectedHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}
