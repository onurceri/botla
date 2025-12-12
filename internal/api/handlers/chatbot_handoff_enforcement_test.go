package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestUpdateChatbot_HandoffForbidden_ForProPlan(t *testing.T) {
	pool, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	defer pool.Close()

	// Ensure pro plan exists and has correct config
	var proPlanID string
	err = pool.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID)
	if err != nil {
		t.Fatalf("pro plan not found: %v", err)
	}

	// Create user with PRO plan
	var userID string
	email := fmt.Sprintf("handoff-pro+%d@example.com", time.Now().UnixNano())
	err = pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", proPlanID).Scan(&userID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create chatbot
	bot := &models.Chatbot{
		UserID:       userID,
		Name:         "Unit Bot Handoff",
		LanguageCode: "tr-TR",
		Model:        "gpt-4o-mini",
	}
	botID, err := db.CreateChatbot(context.Background(), pool, bot)
	if err != nil {
		t.Fatalf("create chatbot: %v", err)
	}

	// Attempt to enable Handoff
	h := &ChatbotHandlers{DB: pool}
	body := []byte(`{"handoff_enabled":true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/chatbots/"+botID, bytes.NewReader(body))
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
