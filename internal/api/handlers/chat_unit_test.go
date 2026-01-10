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

	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func TestChat_NoInfoFound(t *testing.T) {
	t.Parallel()
	oai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "fail", http.StatusInternalServerError) }))
	defer oai.Close()
	qd := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "fail", http.StatusInternalServerError) }))
	defer qd.Close()
	dbx := testdb.OpenTestDB(t)
	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("chatuniq+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	// Create a dummy config and factory to avoid panic
	cfg := &config.Config{OPENAI_API_KEY: "k", OPENAI_API_BASE: oai.URL}
	factory := rag.NewClientFactory(cfg)

	planRepo := repository.NewPostgresPlanRepo(dbx, nil)
	conversationRepo := repository.NewPostgresConversationRepo(dbx)
	analyticsRepo := repository.NewPostgresAnalyticsRepo(dbx)
	chatbotRepo := repository.NewPostgresChatbotRepo(dbx)
	usageRepo := repository.NewPostgresUsageRepo(dbx)

	h := &ChatbotHandlers{DB: dbx, ChatbotRepo: chatbotRepo, PlanRepo: planRepo}
	chatSvc := services.NewChatService(planRepo, conversationRepo, analyticsRepo, nil, nil, nil, factory, nil, nil, usageRepo, logger.New("info")) // factory provided
	chatSvc.SyncAnalytics = true                                                                                                                   // Run analytics synchronously in tests
	ch := &ChatHandlers{DB: dbx, ChatService: chatSvc, Logger: logger.New("info"), ChatbotRepo: chatbotRepo}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}
	body := map[string]any{"name": "Chat Bot", "language": "tr-TR"}
	jb, _ := json.Marshal(body)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	h.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	id := created["id"].(string)
	cr := map[string]any{"message": "selam", "session_id": "s-unit"}
	crb, _ := json.Marshal(cr)
	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+id+"/chat", bytes.NewReader(crb))
	r2.SetPathValue("id", id)
	rr2 := httptest.NewRecorder()
	ch.Chat(rr2, ctx(r2))
	// The mock server returns 500, but the system should gracefully fallback to an empty state message
	if rr2.Code != http.StatusOK {
		t.Fatalf("chat: expected 200, got %d", rr2.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["response"] == "" {
		t.Fatal("chat: expected non-empty response from fallback")
	}
}
