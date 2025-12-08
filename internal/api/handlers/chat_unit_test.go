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
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestChat_NoInfoFound(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	oai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "fail", http.StatusInternalServerError) }))
	defer oai.Close()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	qd := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "fail", http.StatusInternalServerError) }))
	defer qd.Close()
	t.Setenv("QDRANT_URL", qd.URL)
	dbx, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer dbx.Close()
	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("chatuniq+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}
	h := &ChatbotHandlers{DB: dbx}
	chatSvc := services.NewChatService(dbx, nil, nil, nil) // nil clients -> lazy init
	ch := &ChatHandlers{DB: dbx, ChatService: chatSvc}
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
	rr2 := httptest.NewRecorder()
	ch.Chat(rr2, ctx(r2))
	if rr2.Code != http.StatusOK {
		t.Fatalf("chat: %d", rr2.Code)
	}
}
