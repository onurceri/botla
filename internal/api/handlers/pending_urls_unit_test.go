package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onurceri/botla-co/pkg/middleware"
)

func parseBotIDFromPath(path string) (string, bool) {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "chatbots" && i+1 < len(parts) {
			return parts[i+1], true
		}
	}
	return "", false
}

func TestParseBotIDFromPath_PendingURLs(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := parseBotIDFromPath(tt.path)
			if !ok && tt.expected != "" {
				t.Errorf("parseBotIDFromPath(%q) ok = false, want true", tt.path)
			}
			if result != tt.expected {
				t.Errorf("parseBotIDFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
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
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user-123")

	tests := []struct {
		name string
		path string
		want int
	}{
		{
			name: "empty chatbot ID",
			path: "/api/v1/chatbots//pending-urls",
			want: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			h.ListPendingURLs(w, req)

			if w.Code != tc.want {
				t.Errorf("ListPendingURLs with %q: got %d, want %d", tc.path, w.Code, tc.want)
			}
		})
	}
}
