package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
)

func TestSources_StatusAndDelete(t *testing.T) {
	dbx, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer dbx.Close()
	var uid string
	email := fmt.Sprintf("srcuniq+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash) VALUES ($1,$2) RETURNING id`, email, "x").Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}
	ch := &ChatbotHandlers{DB: dbx}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage()}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}
	cb := map[string]any{"name": "Src Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	id := created["id"].(string)
	var mbody bytes.Buffer
	mw := multipart.NewWriter(&mbody)
	mw.WriteField("source_type", "text")
	mw.WriteField("text", "deneme")
	mw.Close()
	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+id+"/sources", bytes.NewReader(mbody.Bytes()))
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))
	if rr2.Code != http.StatusCreated {
		t.Fatalf("create src: %d", rr2.Code)
	}
	var resp map[string]string
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp)
	sid := resp["id"]
	r3 := httptest.NewRequest(http.MethodGet, "/api/v1/sources/"+sid, nil)
	rr3 := httptest.NewRecorder()
	sh.GetSourceStatusOrDelete(rr3, ctx(r3))
	if rr3.Code != http.StatusOK {
		t.Fatalf("status: %d", rr3.Code)
	}
}
