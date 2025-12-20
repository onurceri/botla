package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
)

func TestSources_ETag_Status(t *testing.T) {
	dbx := testdb.OpenTestDB(t)
	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := "etag+" + time.Now().Format("20060102150405") + "@example.com"
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}
	ch := &ChatbotHandlers{DB: dbx}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage()}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	// create chatbot
	cb := map[string]any{"name": "ETag Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	botID := created["id"].(string)

	// add text source
	var mbody bytes.Buffer
	mw := multipart.NewWriter(&mbody)
	_ = mw.WriteField("source_type", "text")
	_ = mw.WriteField("text", "hello world")
	mw.Close()
	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody.Bytes()))
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))
	if rr2.Code != http.StatusCreated {
		t.Fatalf("create src: %d", rr2.Code)
	}
	var resp map[string]string
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp)
	sid := resp["id"]

	// GET status and capture ETag
	r3 := httptest.NewRequest(http.MethodGet, "/api/v1/sources/"+sid, nil)
	rr3 := httptest.NewRecorder()
	sh.GetSourceStatusOrDelete(rr3, ctx(r3))
	if rr3.Code != http.StatusOK {
		t.Fatalf("status1: %d", rr3.Code)
	}
	etag := rr3.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("missing etag")
	}

	// GET with If-None-Match should be 304
	r4 := httptest.NewRequest(http.MethodGet, "/api/v1/sources/"+sid, nil)
	r4.Header.Set("If-None-Match", etag)
	rr4 := httptest.NewRecorder()
	sh.GetSourceStatusOrDelete(rr4, ctx(r4))
	if rr4.Code != http.StatusNotModified {
		t.Fatalf("expected 304, got %d", rr4.Code)
	}
}
