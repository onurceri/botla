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

	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestChatbot_Update_DiscoveryMode_Forbidden_OnZeroCrawlLimit(t *testing.T) {
	db := testdb.OpenTestDB(t)
	defer db.Close()

	var freePlanID string
	if err := db.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if _, err := db.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config,'{}'::jsonb), '{scraping,max_pages_per_crawl}', '0') WHERE id=$1`, freePlanID); err != nil {
		t.Fatalf("update plan: %v", err)
	}

	var uid string
	email := fmt.Sprintf("disc_enf+%d@example.com", time.Now().UnixNano())
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	h := &ChatbotHandlers{DB: db}
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
