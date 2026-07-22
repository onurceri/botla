package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/langconfig"
)

const testIsPublicBotID = "11111111-1111-1111-1111-111111111111"

func botRepoWithVisibility(isPublic bool) *repository.MockChatbotRepo {
	m := repository.NewMockChatbotRepo()
	m.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, IsPublic: isPublic}, nil
	}
	return m
}

func TestPublicChatbotConfig_NonPublicBot_Returns404(t *testing.T) {
	handler := PublicChatbotConfig(botRepoWithVisibility(false))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/"+testIsPublicBotID, nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-public bot, got %d", rr.Code)
	}
}

func TestPublicChatbotConfig_PublicBot_Returns200(t *testing.T) {
	handler := PublicChatbotConfig(botRepoWithVisibility(true))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/"+testIsPublicBotID, nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for public bot, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

func TestPublicChat_NonPublicBot_Returns404(t *testing.T) {
	h := &PublicHandlers{ChatbotRepo: botRepoWithVisibility(false)}
	body := []byte(`{"message":"hi","session_id":"s1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/chatbots/"+testIsPublicBotID+"/chat", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.PublicChat(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-public bot, got %d", rr.Code)
	}
}

func TestPublicRequestHandoff_NonPublicBot_Returns404(t *testing.T) {
	h := &HandoffHandlers{ChatbotRepo: botRepoWithVisibility(false)}
	body := []byte(`{"session_id":"s1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/chatbots/"+testIsPublicBotID+"/handoff", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.PublicRequestHandoff(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-public bot, got %d", rr.Code)
	}
}

func TestPublicSubmitEmail_NonPublicBot_Returns404(t *testing.T) {
	h := &HandoffHandlers{ChatbotRepo: botRepoWithVisibility(false)}
	body := []byte(`{"email":"user@example.com"}`)
	path := "/api/v1/public/chatbots/" + testIsPublicBotID + "/handoff/22222222-2222-2222-2222-222222222222/contact"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.PublicSubmitEmail(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-public bot, got %d", rr.Code)
	}
}

func TestSubmitFeedback_NonPublicBot_Returns404(t *testing.T) {
	h := &PublicHandlers{ChatbotRepo: botRepoWithVisibility(false)}
	body := []byte(`{"message_id":"33333333-3333-3333-3333-333333333333","thumbs_up":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/chatbots/"+testIsPublicBotID+"/feedback", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.SubmitFeedback(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-public bot, got %d", rr.Code)
	}
}

func TestBuildNewChatbot_IsPublicDefaultsTrue(t *testing.T) {
	h := &ChatbotHandlers{}
	bot := h.buildNewChatbot("user-1", nil, nil, createChatbotRequest{Name: "Bot"}, "en-US", langconfig.Get("en"))
	if !bot.IsPublic {
		t.Fatal("expected IsPublic to default to true when omitted from create request")
	}
}

func TestBuildNewChatbot_IsPublicExplicitFalse(t *testing.T) {
	h := &ChatbotHandlers{}
	isPublic := false
	req := createChatbotRequest{Name: "Bot", IsPublic: &isPublic}
	bot := h.buildNewChatbot("user-1", nil, nil, req, "en-US", langconfig.Get("en"))
	if bot.IsPublic {
		t.Fatal("expected IsPublic=false when explicitly set in create request")
	}
}
