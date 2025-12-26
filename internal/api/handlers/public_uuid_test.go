package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicChatbotConfig_InvalidUUID_Returns400(t *testing.T) {
	h := PublicChatbotConfig(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/not-a-uuid", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid ID format") {
		t.Fatalf("expected body to contain %q, got %s", "Invalid ID format", rr.Body.String())
	}
}

func TestPublicChat_InvalidUUID_Returns400(t *testing.T) {
	h := &PublicHandlers{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/chatbots/not-a-uuid/chat", nil)
	rr := httptest.NewRecorder()

	h.PublicChat(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid ID format") {
		t.Fatalf("expected body to contain %q, got %s", "Invalid ID format", rr.Body.String())
	}
}

func TestPublicSubmitFeedback_InvalidUUID_Returns400(t *testing.T) {
	h := &PublicHandlers{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/chatbots/not-a-uuid/feedback", nil)
	rr := httptest.NewRecorder()

	h.SubmitFeedback(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid ID format") {
		t.Fatalf("expected body to contain %q, got %s", "Invalid ID format", rr.Body.String())
	}
}
