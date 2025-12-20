package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
)

// TestChatbotSources_Unauthorized tests unauthenticated requests
func TestChatbotSources_Unauthorized(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}

	r := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/abc/sources", nil)
	rr := httptest.NewRecorder()
	sh.ChatbotSources(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ChatbotSources() status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestChatbotSources_InvalidPath tests invalid path handling
func TestChatbotSources_InvalidPath(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user-123")

	tests := []struct {
		path string
		want int
	}{
		{"/api/v1/chatbots//sources", http.StatusNotFound},
		{"/api/v1/chatbots/new/sources", http.StatusBadRequest},
	}

	for _, tc := range tests {
		r := httptest.NewRequest(http.MethodGet, tc.path, nil)
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()
		sh.ChatbotSources(rr, r)

		if rr.Code != tc.want {
			t.Errorf("ChatbotSources(%q) status = %d, want %d", tc.path, rr.Code, tc.want)
		}
	}
}

// TestChatbotSources_MethodNotAllowed tests unsupported HTTP methods
func TestChatbotSources_MethodNotAllowed(t *testing.T) {
	dbx := testdb.OpenTestDB(t)

	// Create test user and chatbot
	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("method_test+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	var botID string
	if err := dbx.QueryRow(`INSERT INTO chatbots (user_id, name, system_prompt, language_id, model, 
		theme_color, welcome_message, position, bot_message_color, user_message_color, 
		bot_message_text_color, user_message_text_color, chat_font_family, chat_header_color, 
		chat_header_text_color, chat_background_color) VALUES ($1, $2, '', (SELECT id FROM languages WHERE code='tr-TR'), 
		'gpt-4o-mini', '#3b82f6', 'Hello', 'bottom-right', '#f3f4f6', '#3b82f6', '#1f2937', '#ffffff', 
		'Inter', '#3b82f6', '#ffffff', '#ffffff') RETURNING id`, uid, "Method Test Bot").Scan(&botID); err != nil {
		t.Fatalf("chatbot: %v", err)
	}

	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage()}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, uid)

	// Test unsupported methods
	for _, method := range []string{http.MethodPut, http.MethodPatch, http.MethodDelete} {
		r := httptest.NewRequest(method, "/api/v1/chatbots/"+botID+"/sources", nil)
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()
		sh.ChatbotSources(rr, r)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("ChatbotSources(%s) status = %d, want %d", method, rr.Code, http.StatusMethodNotAllowed)
		}
	}
}
