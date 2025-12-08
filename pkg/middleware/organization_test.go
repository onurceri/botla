package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractTenantContext_NoHeaders(t *testing.T) {
	mw := ExtractTenantContext()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots", nil)

	var gotOrgID, gotWsID string
	var orgOK, wsOK bool

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotOrgID, orgOK = OrgIDFromContext(r.Context())
		gotWsID, wsOK = WorkspaceIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if orgOK || gotOrgID != "" {
		t.Errorf("expected no org ID, got %q (ok=%v)", gotOrgID, orgOK)
	}
	if wsOK || gotWsID != "" {
		t.Errorf("expected no workspace ID, got %q (ok=%v)", gotWsID, wsOK)
	}
}

func TestExtractTenantContext_WithOrgHeader(t *testing.T) {
	mw := ExtractTenantContext()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots", nil)
	req.Header.Set(HeaderOrgID, "org-123")

	var gotOrgID string
	var orgOK bool

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotOrgID, orgOK = OrgIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if !orgOK || gotOrgID != "org-123" {
		t.Errorf("expected org-123, got %q (ok=%v)", gotOrgID, orgOK)
	}
}

func TestExtractTenantContext_WithWorkspaceHeader(t *testing.T) {
	mw := ExtractTenantContext()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots", nil)
	req.Header.Set(HeaderWorkspaceID, "ws-456")

	var gotWsID string
	var wsOK bool

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotWsID, wsOK = WorkspaceIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if !wsOK || gotWsID != "ws-456" {
		t.Errorf("expected ws-456, got %q (ok=%v)", gotWsID, wsOK)
	}
}

func TestExtractTenantContext_BothHeaders(t *testing.T) {
	mw := ExtractTenantContext()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots", nil)
	req.Header.Set(HeaderOrgID, "org-abc")
	req.Header.Set(HeaderWorkspaceID, "ws-xyz")

	var gotOrgID, gotWsID string
	var orgOK, wsOK bool

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotOrgID, orgOK = OrgIDFromContext(r.Context())
		gotWsID, wsOK = WorkspaceIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if !orgOK || gotOrgID != "org-abc" {
		t.Errorf("expected org-abc, got %q (ok=%v)", gotOrgID, orgOK)
	}
	if !wsOK || gotWsID != "ws-xyz" {
		t.Errorf("expected ws-xyz, got %q (ok=%v)", gotWsID, wsOK)
	}
}

func TestExtractOrgIDFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/organizations/abc/workspaces", "abc"},
		{"/api/v1/organizations/123", "123"},
		{"/api/v1/chatbots", ""},
		{"/api/v1/organizations/", ""},
		{"/organizations/xyz/test", "xyz"},
	}

	for _, tt := range tests {
		got := extractOrgIDFromPath(tt.path)
		if got != tt.expected {
			t.Errorf("extractOrgIDFromPath(%q) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}

func TestHasMinRole(t *testing.T) {
	tests := []struct {
		userRole string
		minRole  string
		expected bool
	}{
		{"owner", "member", true},
		{"owner", "admin", true},
		{"owner", "owner", true},
		{"admin", "member", true},
		{"admin", "admin", true},
		{"admin", "owner", false},
		{"member", "member", true},
		{"member", "admin", false},
		{"member", "owner", false},
		{"", "member", false},
	}

	for _, tt := range tests {
		got := hasMinRole(tt.userRole, tt.minRole)
		if got != tt.expected {
			t.Errorf("hasMinRole(%q, %q) = %v, want %v", tt.userRole, tt.minRole, got, tt.expected)
		}
	}
}
