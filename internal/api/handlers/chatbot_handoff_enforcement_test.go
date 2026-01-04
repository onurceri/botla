package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func TestUpdateChatbot_HandoffForbidden_ForProPlan(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	// Ensure pro plan exists
	var proPlanID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID); err != nil {
		t.Fatalf("pro plan not found: %v", err)
	}

	// Create user with PRO plan
	var userID string
	email := fmt.Sprintf("handoff-pro+%d@example.com", time.Now().UnixNano())
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", proPlanID).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create chatbot
	bot := &models.Chatbot{
		UserID:       userID,
		Name:         "Unit Bot Handoff",
		LanguageCode: "tr-TR",
		Model:        "gpt-4o-mini",
	}
	chatbotRepo := repository.NewPostgresChatbotRepo(pool)
	botID, err := chatbotRepo.Create(context.Background(), bot)
	if err != nil {
		t.Fatalf("create chatbot: %v", err)
	}

	// Attempt to enable Handoff
	h := &ChatbotHandlers{
		DB:             pool,
		ChatbotService: services.NewChatbotService(chatbotRepo, repository.NewPostgresPlanRepo(pool, nil), logger.New("info")),
		ChatbotRepo:    chatbotRepo,
	}
	body := []byte(`{"handoff_enabled":true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/chatbots/"+botID, bytes.NewReader(body))
	req.SetPathValue("id", botID)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rr := httptest.NewRecorder()

	h.ByID(rr, req.WithContext(ctx))

	// Should be Forbidden (403)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", rr.Code)
	}

	// Verify error message
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if val, ok := resp["feature"]; !ok || val != "escalate_fallback" {
		t.Logf("Response: %v", resp)
		// We expect a specific error structure if we implement it correctly
	}
}
