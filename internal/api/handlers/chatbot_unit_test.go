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
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestChatbot_ListCreateUpdateDelete(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	var uid string
	var freePlanID string
	if err := db.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("botuniq+%d@example.com", time.Now().UnixNano())
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}
	h := &ChatbotHandlers{DB: db}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}
	// create
	body := map[string]any{"name": "Unit Bot", "language": "en-US", "suggestions_enabled": true, "suggested_questions": []string{"A"}}
	jb, _ := json.Marshal(body)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	h.ListOrCreate(rr1, ctx(r1))
	if rr1.Code != http.StatusCreated {
		t.Fatalf("create: %d", rr1.Code)
	}
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	id := created["id"].(string)
	// list
	r2 := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots", nil)
	rr2 := httptest.NewRecorder()
	h.ListOrCreate(rr2, ctx(r2))
	if rr2.Code != http.StatusOK {
		t.Fatalf("list: %d", rr2.Code)
	}
	// get
	r3 := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+id, nil)
	rr3 := httptest.NewRecorder()
	h.ByID(rr3, ctx(r3))
	if rr3.Code != http.StatusOK {
		t.Fatalf("get: %d", rr3.Code)
	}
	// update
	ub := map[string]any{"name": "Renamed"}
	uj, _ := json.Marshal(ub)
	r4 := httptest.NewRequest(http.MethodPut, "/api/v1/chatbots/"+id, bytes.NewReader(uj))
	rr4 := httptest.NewRecorder()
	h.ByID(rr4, ctx(r4))
	if rr4.Code != http.StatusOK {
		t.Fatalf("upd: %d", rr4.Code)
	}
	// delete
	r5 := httptest.NewRequest(http.MethodDelete, "/api/v1/chatbots/"+id, nil)
	rr5 := httptest.NewRecorder()
	h.ByID(rr5, ctx(r5))
	if rr5.Code != http.StatusNoContent {
		t.Fatalf("del: %d", rr5.Code)
	}
}
