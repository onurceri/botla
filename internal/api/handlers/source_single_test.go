package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/pkg/middleware"
	"github.com/onurceri/botla-app/pkg/storage"
)

// TestGetSourceStatusOrDelete_Unauthorized tests unauthenticated source access
func TestGetSourceStatusOrDelete_Unauthorized(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}

	r := httptest.NewRequest(http.MethodGet, "/api/v1/sources/abc", nil)
	rr := httptest.NewRecorder()
	sh.GetSourceStatusOrDelete(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("GetSourceStatusOrDelete() status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestGetSourceStatusOrDelete_InvalidPath tests invalid source paths
func TestGetSourceStatusOrDelete_InvalidPath(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user-123")

	tests := []struct {
		path string
		want int
	}{
		{"/api/v1/sources/", http.StatusNotFound},
		{"/api/v1/sources/abc/refresh", http.StatusNotFound},
	}

	for _, tc := range tests {
		r := httptest.NewRequest(http.MethodGet, tc.path, nil)
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()
		sh.GetSourceStatusOrDelete(rr, r)

		if rr.Code != tc.want {
			t.Errorf("GetSourceStatusOrDelete(%q) status = %d, want %d", tc.path, rr.Code, tc.want)
		}
	}
}
