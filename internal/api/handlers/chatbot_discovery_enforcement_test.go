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

	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
)


func TestChatbot_Update_DiscoveryMode_Forbidden_OnZeroCrawlLimit(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)

	var freePlanID string
	if err := dbConn.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	// Use testdb.UpdatePlanLimit to update the plan limit (new pattern)
	if err := testdb.UpdatePlanLimit(context.Background(), dbConn, "free", "scraping_max_pages_per_crawl", 0); err != nil {
		t.Fatalf("update plan: %v", err)
	}

	var uid string
	email := fmt.Sprintf("disc_enf+%d@example.com", time.Now().UnixNano())
	if err := dbConn.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	h := &ChatbotHandlers{
		DB:             dbConn,
		ChatbotService: services.NewChatbotService(repository.NewPostgresChatbotRepo(dbConn), repository.NewPostgresPlanRepo(dbConn, nil), logger.New("info")),
		ChatbotRepo:    repository.NewPostgresChatbotRepo(dbConn),
		PlanRepo:       repository.NewPostgresPlanRepo(dbConn, nil),
	}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	cb := map[string]any{"name": "Disc Enf Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	h.ListOrCreate(rr1, ctx(r1))
	if rr1.Code != http.StatusCreated {
		t.Fatalf("create: %d", rr1.Code)
	}
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	botID := created["id"].(string)

	update := map[string]any{"discovery_mode": "auto"}
	uj, _ := json.Marshal(update)
	r2 := httptest.NewRequest(http.MethodPut, "/api/v1/chatbots/"+botID, bytes.NewReader(uj))
	r2.SetPathValue("id", botID)
	rr2 := httptest.NewRecorder()
	h.ByID(rr2, ctx(r2))
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("update: got %d want %d", rr2.Code, http.StatusForbidden)
	}
	var resp map[string]any
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp)
	if v, ok := resp["upgrade_required"].(bool); !ok || !v {
		t.Fatalf("upgrade_required missing or false: %v", resp)
	}
}
