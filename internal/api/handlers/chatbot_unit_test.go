package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestUpdateChatbot_AutoRefreshForbidden_ForFreePlan(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	var freePlanID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("free plan: %v", err)
	}

	var userID string
	email := fmt.Sprintf("auto-refresh-free+%d@example.com", time.Now().UnixNano())
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	bot := &models.Chatbot{
		UserID:       userID,
		Name:         "Unit Bot",
		LanguageCode: "tr-TR",
		Model:        "gpt-4o-mini",
	}
	var botID string
	var err error
	botID, err = db.CreateChatbot(context.Background(), pool, bot)
	if err != nil {
		t.Fatalf("create chatbot: %v", err)
	}

	h := &ChatbotHandlers{
		DB:             pool,
		ChatbotService: services.NewChatbotService(pool, nil),
	}
	body := []byte(`{"refresh_policy":"auto","refresh_frequency":"weekly"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/chatbots/"+botID, bytes.NewReader(body))
	req.SetPathValue("id", botID)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rr := httptest.NewRecorder()

	h.ByID(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", rr.Code)
	}
}

func TestUpdateChatbot_SecureEmbedForbidden_ForFreePlan(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	var freePlanID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("free plan: %v", err)
	}

	var userID string
	email := fmt.Sprintf("secure-embed-free+%d@example.com", time.Now().UnixNano())
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	bot := &models.Chatbot{
		UserID:       userID,
		Name:         "Unit Bot",
		LanguageCode: "tr-TR",
		Model:        "gpt-4o-mini",
	}
	var botID string
	var err error
	botID, err = db.CreateChatbot(context.Background(), pool, bot)
	if err != nil {
		t.Fatalf("create chatbot: %v", err)
	}

	h := &ChatbotHandlers{
		DB:             pool,
		ChatbotService: services.NewChatbotService(pool, nil),
	}
	body := []byte(`{"secure_embed_enabled": true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/chatbots/"+botID, bytes.NewReader(body))
	req.SetPathValue("id", botID)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rr := httptest.NewRecorder()

	h.ByID(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", rr.Code)
	}
}
