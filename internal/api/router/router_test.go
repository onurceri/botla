package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/services"
)

func TestRegisterAuthRoutes(t *testing.T) {
	mux := http.NewServeMux()
	ah := &handlers.AuthHandlers{} // Zero value is fine for routing checks if we hit early returns
	registerAuthRoutes(mux, ah, "secret")

	tests := []struct {
		name       string
		method     string
		path       string
		wantCode   int
		wantHeader bool // If true, checks if response is NOT 404
	}{
		{"Register_MethodNotAllowed", http.MethodGet, "/api/v1/auth/register", http.StatusMethodNotAllowed, false},
		{"Login_MethodNotAllowed", http.MethodGet, "/api/v1/auth/login", http.StatusMethodNotAllowed, false},
		{"Refresh_MethodNotAllowed", http.MethodGet, "/api/v1/auth/refresh", http.StatusMethodNotAllowed, false},
		{"Logout_MethodNotAllowed", http.MethodGet, "/api/v1/auth/logout", http.StatusMethodNotAllowed, false},
		{"Protected_Unauthorized", http.MethodGet, "/api/v1/protected", http.StatusUnauthorized, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if tt.wantHeader {
				if w.Code == http.StatusNotFound {
					t.Errorf("path %s not registered (got 404)", tt.path)
				}
			} else {
				if w.Code != tt.wantCode {
					t.Errorf("path %s: got %d want %d", tt.path, w.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestRegisterPublicRoutes(t *testing.T) {
	mux := http.NewServeMux()
	// Initialize handlers with minimal/nil deps.
	// Note: handlers logic might panic if we go deep, but we aim for early returns or routing checks.
	hoh := &handlers.HandoffHandlers{}
	ph := &handlers.PublicHandlers{}
	// We don't need a real DB for routing check if we don't trigger DB calls.
	// However, PublicChatbotConfig handler uses pool directly. It returns a closure.
	// `handlers.PublicChatbotConfig(pool)(w, r)`
	// If pool is nil, and the handler tries to use it immediately, it might panic.
	// Let's check PublicChatbotConfig implementation or avoid hitting that path deeply.
	// Actually, PublicChatbotConfig is the fallback.

	registerPublicRoutes(mux, "secret", hoh, ph, nil)

	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
	}{
		// Use GET for POST-only endpoints to trigger 405 and avoid nil DB panic
		{"Handoff_Request_MethodNotAllowed", http.MethodGet, "/api/v1/public/chatbots/bot1/handoff", http.StatusMethodNotAllowed},
		{"Handoff_Contact_MethodNotAllowed", http.MethodGet, "/api/v1/public/chatbots/bot1/handoff/req1/contact", http.StatusMethodNotAllowed},
		// For Chat and Feedback, if they don't check method early, we might panic.
		// Let's assume they do. If not, we'll see a panic and fix.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("path %s: got %d want %d", tt.path, w.Code, tt.wantCode)
			}
		})
	}
}

func TestRegisterSourceRoutes(t *testing.T) {
	mux := http.NewServeMux()
	sh := &handlers.SourcesHandlers{}
	RegisterSourceRoutes(mux, "secret", sh)

	// Should be protected by middleware -> 401
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources/123", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for protected source route, got %d", w.Code)
	}
}

func TestRegisterOrgRoutes(t *testing.T) {
	mux := http.NewServeMux()
	orgSvc := &services.OrganizationService{}
	oh := &handlers.OrganizationHandlers{}
	wh := &handlers.WorkspaceHandlers{}

	registerOrgRoutes(mux, "secret", orgSvc, oh, wh)

	// All org routes are protected -> 401 (or 405 if method mismatch)
	tests := []struct {
		method   string
		path     string
		wantCode int
	}{
		{http.MethodGet, "/api/v1/organizations", http.StatusUnauthorized},
		{http.MethodPatch, "/api/v1/organizations/1", http.StatusUnauthorized}, // PATCH is registered
		{http.MethodGet, "/api/v1/organizations/1/workspaces", http.StatusUnauthorized},
		{http.MethodGet, "/api/v1/organizations/1/members", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != tt.wantCode {
			t.Errorf("method %s path %s: expected %d, got %d", tt.method, tt.path, tt.wantCode, w.Code)
		}
	}
}

func TestChatbotsDispatchHandler_Routes(t *testing.T) {
	ch := &handlers.ChatbotHandlers{}
	sh := &handlers.SourcesHandlers{}
	chh := &handlers.ChatHandlers{}
	puh := &handlers.PendingURLsHandlers{}
	acth := &handlers.ActionHandlers{}
	hoh := &handlers.HandoffHandlers{}
	anh := &handlers.AnalyticsHandlers{}
	sugh := &handlers.SuggestionsHandlers{}

	// Use nil for the remaining arguments if any
	h := ChatbotsDispatchHandler("secret", ch, sh, chh, puh, acth, hoh, anh, sugh)

	// Since everything is protected by AuthMiddleware, we expect 401 for all valid routes if no token provided.
	// If a route is NOT registered, we might get 404 (if it falls through) OR 401 (if the whole handler is wrapped).
	// The dispatch handler wraps the *entire* mux with AuthMiddleware.
	// So ANY request to it should return 401 (Unauthorized), NOT 404.
	// If the middleware wasn't applied, we might get 404 from the internal mux.

	cases := []struct {
		path string
		code int
	}{
		{"/api/v1/chatbots/x/sources", http.StatusUnauthorized},
		{"/api/v1/chatbots/x/analytics/sources", http.StatusUnauthorized},
		{"/api/v1/chatbots/x/chat", http.StatusUnauthorized},
		{"/api/v1/chatbots/x", http.StatusUnauthorized},
		{"/api/v1/chatbots/x/pending-urls", http.StatusUnauthorized},
		{"/api/v1/chatbots/x/actions", http.StatusUnauthorized},
		{"/api/v1/chatbots/x/handoff-requests", http.StatusUnauthorized},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != c.code {
			t.Fatalf("path %s: got %d want %d", c.path, w.Code, c.code)
		}
	}
}
