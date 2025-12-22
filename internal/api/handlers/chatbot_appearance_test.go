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

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestUpdateAppearance_NewFields(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	// Create user
	var proPlanID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID); err != nil {
		t.Fatalf("pro plan: %v", err)
	}

	var userID string
	email := fmt.Sprintf("appearance-test+%d@example.com", time.Now().UnixNano())
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", proPlanID).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create chatbot
	bot := &models.Chatbot{
		UserID:       userID,
		Name:         "Appearance Bot",
		LanguageCode: "tr-TR",
		Model:        "gpt-4o-mini",
	}
	botID, err := db.CreateChatbot(context.Background(), pool, bot)
	if err != nil {
		t.Fatalf("create chatbot: %v", err)
	}

	h := &ChatbotHandlers{
		DB:             pool,
		ChatbotService: services.NewChatbotService(pool, nil),
	}

	// Update new fields
	payload := map[string]interface{}{
		"bubble_radius":          "16px",
		"input_background_color": "#f0f0f0",
		"input_text_color":       "#333333",
		"send_button_color":      "#ff0000",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/chatbots/"+botID+"/appearance", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rr := httptest.NewRecorder()

	h.UpdateAppearance(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var updatedBot models.Chatbot
	err = json.NewDecoder(rr.Body).Decode(&updatedBot)
	if err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if updatedBot.BubbleRadius != "16px" {
		t.Errorf("expected BubbleRadius '16px', got '%s'", updatedBot.BubbleRadius)
	}
	if updatedBot.InputBackgroundColor != "#f0f0f0" {
		t.Errorf("expected InputBackgroundColor '#f0f0f0', got '%s'", updatedBot.InputBackgroundColor)
	}
	if updatedBot.InputTextColor != "#333333" {
		t.Errorf("expected InputTextColor '#333333', got '%s'", updatedBot.InputTextColor)
	}
	if updatedBot.SendButtonColor != "#ff0000" {
		t.Errorf("expected SendButtonColor '#ff0000', got '%s'", updatedBot.SendButtonColor)
	}

	// Verify persistence
	persisted, err := db.GetChatbotByID(context.Background(), pool, botID)
	if err != nil {
		t.Fatalf("get chatbot: %v", err)
	}
	if persisted.BubbleRadius != "16px" {
		t.Errorf("persisted BubbleRadius mismatch")
	}
}

func TestPublicChatbotConfig_NewFields(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	// Create user
	var proPlanID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID); err != nil {
		t.Fatalf("pro plan: %v", err)
	}

	var userID string
	email := fmt.Sprintf("public-config-test+%d@example.com", time.Now().UnixNano())
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", proPlanID).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create chatbot with new fields populated
	bot := &models.Chatbot{
		UserID:               userID,
		Name:                 "Public Config Bot",
		LanguageCode:         "en-US",
		Model:                "gpt-4o-mini",
		BubbleRadius:         "12px",
		InputBackgroundColor: "#abcdef",
		InputTextColor:       "#123456",
		SendButtonColor:      "#654321",
	}
	botID, err := db.CreateChatbot(context.Background(), pool, bot)
	if err != nil {
		t.Fatalf("create chatbot: %v", err)
	}

	// Request public config
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/"+botID, nil)
	rr := httptest.NewRecorder()

	// Handler under test
	handler := PublicChatbotConfig(pool)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var publicConf publicChatbot
	if err := json.NewDecoder(rr.Body).Decode(&publicConf); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if publicConf.BubbleRadius != "12px" {
		t.Errorf("expected BubbleRadius '12px', got '%s'", publicConf.BubbleRadius)
	}
	if publicConf.InputBackgroundColor != "#abcdef" {
		t.Errorf("expected InputBackgroundColor '#abcdef', got '%s'", publicConf.InputBackgroundColor)
	}
	if publicConf.InputTextColor != "#123456" {
		t.Errorf("expected InputTextColor '#123456', got '%s'", publicConf.InputTextColor)
	}
	if publicConf.SendButtonColor != "#654321" {
		t.Errorf("expected SendButtonColor '#654321', got '%s'", publicConf.SendButtonColor)
	}
}
