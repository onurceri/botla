package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractChatbotIDFromPendingURLPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "basic pending-urls path",
			path:     "/api/v1/chatbots/abc123/pending-urls",
			expected: "abc123",
		},
		{
			name:     "pending-urls approve path",
			path:     "/api/v1/chatbots/xyz789/pending-urls/approve",
			expected: "xyz789",
		},
		{
			name:     "pending-urls reject path",
			path:     "/api/v1/chatbots/test-id/pending-urls/reject",
			expected: "test-id",
		},
		{
			name:     "pending-urls clear path",
			path:     "/api/v1/chatbots/uuid-here/pending-urls/clear",
			expected: "uuid-here",
		},
		{
			name:     "invalid path - no prefix",
			path:     "/other/path/chatbots/abc/pending-urls",
			expected: "",
		},
		{
			name:     "empty after prefix",
			path:     "/api/v1/chatbots/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractChatbotIDFromPendingURLPath(tt.path)
			if result != tt.expected {
				t.Errorf("extractChatbotIDFromPendingURLPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestListPendingURLs_MethodNotAllowed(t *testing.T) {
	h := &PendingURLsHandlers{}

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/chatbots/abc/pending-urls", nil)
			w := httptest.NewRecorder()

			h.ListPendingURLs(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("ListPendingURLs with %s: got %d, want %d", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestApprovePendingURLs_MethodNotAllowed(t *testing.T) {
	h := &PendingURLsHandlers{}

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/chatbots/abc/pending-urls/approve", nil)
			w := httptest.NewRecorder()

			h.ApprovePendingURLs(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("ApprovePendingURLs with %s: got %d, want %d", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestRejectPendingURLs_MethodNotAllowed(t *testing.T) {
	h := &PendingURLsHandlers{}

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/chatbots/abc/pending-urls/reject", nil)
			w := httptest.NewRecorder()

			h.RejectPendingURLs(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("RejectPendingURLs with %s: got %d, want %d", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestClearPendingURLs_MethodNotAllowed(t *testing.T) {
	h := &PendingURLsHandlers{}

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/chatbots/abc/pending-urls/clear", nil)
			w := httptest.NewRecorder()

			h.ClearPendingURLs(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("ClearPendingURLs with %s: got %d, want %d", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestListPendingURLs_InvalidChatbotID(t *testing.T) {
	h := &PendingURLsHandlers{}

	// Path without chatbot ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots//pending-urls", nil)
	w := httptest.NewRecorder()

	h.ListPendingURLs(w, req)

	// Should return bad request for empty chatbot ID
	if w.Code != http.StatusBadRequest {
		t.Errorf("ListPendingURLs with empty ID: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}
