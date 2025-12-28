package guards

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestRequirePlatformAdmin(t *testing.T) {
	tests := []struct {
		name           string
		isAdmin        bool
		hasUser        bool
		expectedStatus int
	}{
		{
			name:           "Admin user allowed",
			isAdmin:        true,
			hasUser:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-admin user forbidden",
			isAdmin:        false,
			hasUser:        true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "No user unauthorized",
			isAdmin:        false,
			hasUser:        false,
			expectedStatus: http.StatusForbidden, // Current implementation returns 403 if flag is missing/false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequirePlatformAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.hasUser {
				ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, "user-123")
				ctx = context.WithValue(ctx, middleware.ContextKeyIsPlatformAdmin, tt.isAdmin)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
