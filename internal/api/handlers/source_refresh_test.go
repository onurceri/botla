package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/pkg/storage"
)

// TestRefreshSource_Unauthorized tests unauthenticated refresh
func TestRefreshSource_Unauthorized(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}
	
	r := httptest.NewRequest(http.MethodPost, "/api/v1/sources/abc/refresh", nil)
	rr := httptest.NewRecorder()
	sh.RefreshSource(rr, r)
	
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RefreshSource() status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestRefreshSource_MethodNotAllowed tests non-POST refresh requests
func TestRefreshSource_MethodNotAllowed(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}
	
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		r := httptest.NewRequest(method, "/api/v1/sources/abc/refresh", nil)
		rr := httptest.NewRecorder()
		sh.RefreshSource(rr, r)
		
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("RefreshSource(%s) status = %d, want %d", method, rr.Code, http.StatusMethodNotAllowed)
		}
	}
}
