package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesNewID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request ID is in context
		reqID := RequestIDFromContext(r.Context())
		if reqID == "" {
			t.Error("expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify request ID is in response header
	respID := rec.Header().Get(RequestIDHeader)
	if respID == "" {
		t.Error("expected request ID in response header")
	}
	
	// Verify it looks like a UUID
	if len(respID) != 36 {
		t.Errorf("expected UUID format, got %s", respID)
	}
}

func TestRequestID_UsesExistingHeader(t *testing.T) {
	existingID := "existing-request-id-12345"
	
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := RequestIDFromContext(r.Context())
		if reqID != existingID {
			t.Errorf("expected %s, got %s", existingID, reqID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify same ID is returned
	respID := rec.Header().Get(RequestIDHeader)
	if respID != existingID {
		t.Errorf("expected %s in response, got %s", existingID, respID)
	}
}

func TestRequestIDFromContext_NoID(t *testing.T) {
	ctx := context.Background()
	id := RequestIDFromContext(ctx)
	if id != "" {
		t.Errorf("expected empty string, got %s", id)
	}
}
