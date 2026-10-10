package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
)

func TestPublicChatbotConfig_IncludesLanguageCode(t *testing.T) {
	repo := repository.NewMockChatbotRepo()
	repo.GetByIDFunc = func(ctx context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, IsPublic: true, LanguageCode: "en-US"}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/"+testIsPublicBotID, nil)
	rr := httptest.NewRecorder()

	PublicChatbotConfig(repo).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["language_code"] != "en-US" {
		t.Fatalf("expected language_code en-US, got %v", body["language_code"])
	}
}
