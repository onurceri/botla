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

// TestParseChatbotIDFromPath tests chatbot ID extraction from path
func TestParseChatbotIDFromPath(t *testing.T) {
	tests := []struct {
		path    string
		wantID  string
		wantOK  bool
	}{
		{"/api/v1/chatbots/abc/sources", "abc", true},
		{"/api/v1/chatbots/uuid-123/sources", "uuid-123", true},
		{"/api/v1/chatbots//sources", "", false},
		{"/api/v1/chatbots/abc/x", "", false},
		{"/api/v1/chatbots/abc", "", false},
		{"/api/v1/chatbots/", "", false},
		{"/wrong/path", "", false},
		{"/api/v1/chatbots/abc/sources/extra", "", false},
	}
	for _, tc := range tests {
		id, ok := parseChatbotIDFromPath(tc.path)
		if ok != tc.wantOK || id != tc.wantID {
			t.Errorf("parseChatbotIDFromPath(%q) = (%q, %v), want (%q, %v)",
				tc.path, id, ok, tc.wantID, tc.wantOK)
		}
	}
}

// TestParseSourceIDFromPath tests source ID extraction from path
func TestParseSourceIDFromPath(t *testing.T) {
	tests := []struct {
		path    string
		wantID  string
		wantOK  bool
	}{
		{"/api/v1/sources/abc", "abc", true},
		{"/api/v1/sources/uuid-123", "uuid-123", true},
		{"/api/v1/sources/", "", false},
		{"/api/v1/sources/abc/refresh", "", false},
		{"/api/v1/sources/abc/extra", "", false},
		{"/wrong/path", "", false},
	}
	for _, tc := range tests {
		id, ok := parseSourceIDFromPath(tc.path)
		if ok != tc.wantOK || id != tc.wantID {
			t.Errorf("parseSourceIDFromPath(%q) = (%q, %v), want (%q, %v)",
				tc.path, id, ok, tc.wantID, tc.wantOK)
		}
	}
}

// TestParseRefreshSourceIDFromPath tests refresh source ID extraction
func TestParseRefreshSourceIDFromPath(t *testing.T) {
	tests := []struct {
		path    string
		wantID  string
		wantOK  bool
	}{
		{"/api/v1/sources/abc/refresh", "abc", true},
		{"/api/v1/sources/uuid-123/refresh", "uuid-123", true},
		{"/api/v1/sources//refresh", "", false},
		{"/api/v1/sources/abc", "", false},
		{"/api/v1/sources/abc/other", "", false},
		{"/wrong/path/refresh", "", false},
	}
	for _, tc := range tests {
		id, ok := parseRefreshSourceIDFromPath(tc.path)
		if ok != tc.wantOK || id != tc.wantID {
			t.Errorf("parseRefreshSourceIDFromPath(%q) = (%q, %v), want (%q, %v)",
				tc.path, id, ok, tc.wantID, tc.wantOK)
		}
	}
}

// TestIsPDFContentType tests PDF content type detection
func TestIsPDFContentType(t *testing.T) {
	tests := []struct {
		ct      string
		name    string
		want    bool
	}{
		{"application/pdf", "x.txt", true},
		{"", "x.pdf", true},
		{"application/pdf", "x.pdf", true},
		{"text/plain", "x.txt", false},
		{"application/octet-stream", "x.doc", false},
		{"", "", false},
	}
	for _, tc := range tests {
		got := isPDFContentType(tc.ct, tc.name)
		if got != tc.want {
			t.Errorf("isPDFContentType(%q, %q) = %v, want %v",
				tc.ct, tc.name, got, tc.want)
		}
	}
}

// TestComputeHash tests hash computation
func TestComputeHash(t *testing.T) {
	data := []byte("test data")
	hash := computeHash(data)
	if len(hash) != 64 { // SHA256 produces 64 hex chars
		t.Errorf("computeHash() returned hash of length %d, want 64", len(hash))
	}
	// Same input should produce same hash
	hash2 := computeHash(data)
	if hash != hash2 {
		t.Error("computeHash() is not deterministic")
	}
	// Different input should produce different hash
	hash3 := computeHash([]byte("different data"))
	if hash == hash3 {
		t.Error("computeHash() produced same hash for different inputs")
	}
}

// TestQuotaError tests quota error interface
func TestQuotaError(t *testing.T) {
	err := &quotaError{msg: "test error"}
	if err.Error() != "test error" {
		t.Errorf("quotaError.Error() = %q, want %q", err.Error(), "test error")
	}
}

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
	dbx, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Skipf("db not available: %v", err)
	}
	defer dbx.Close()
	
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

// TestGetSourceStatusOrDelete_Unauthorized tests unauthenticated source access
func TestGetSourceStatusOrDelete_Unauthorized(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}
	
	r := httptest.NewRequest(http.MethodGet, "/api/v1/sources/abc", nil)
	rr := httptest.NewRecorder()
	sh.GetSourceStatusOrDelete(rr, r)
	
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("GetSourceStatusOrDelete() status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestGetSourceStatusOrDelete_InvalidPath tests invalid source paths
func TestGetSourceStatusOrDelete_InvalidPath(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user-123")
	
	tests := []struct {
		path string
		want int
	}{
		{"/api/v1/sources/", http.StatusNotFound},
		{"/api/v1/sources/abc/refresh", http.StatusNotFound},
	}
	
	for _, tc := range tests {
		r := httptest.NewRequest(http.MethodGet, tc.path, nil)
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()
		sh.GetSourceStatusOrDelete(rr, r)
		
		if rr.Code != tc.want {
			t.Errorf("GetSourceStatusOrDelete(%q) status = %d, want %d", tc.path, rr.Code, tc.want)
		}
	}
}

// TestRefreshSource_Unauthorized tests unauthenticated refresh
func TestRefreshSource_Unauthorized(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}
	
	r := httptest.NewRequest(http.MethodPost, "/api/v1/sources/abc/refresh", nil)
	rr := httptest.NewRecorder()
	sh.RefreshSource(rr, r)
	
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RefreshSource() status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestRefreshSource_MethodNotAllowed tests non-POST refresh requests
func TestRefreshSource_MethodNotAllowed(t *testing.T) {
	sh := &SourcesHandlers{DB: nil, Storage: storage.NewMemoryStorage()}
	
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		r := httptest.NewRequest(method, "/api/v1/sources/abc/refresh", nil)
		rr := httptest.NewRecorder()
		sh.RefreshSource(rr, r)
		
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("RefreshSource(%s) status = %d, want %d", method, rr.Code, http.StatusMethodNotAllowed)
		}
	}
}

// TestSources_TextCreation tests text source creation end-to-end
func TestSources_TextCreation(t *testing.T) {
	dbx, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Skipf("db not available: %v", err)
	}
	defer dbx.Close()

	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("text_create+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	ch := &ChatbotHandlers{DB: dbx}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage()}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	// Create chatbot
	cb := map[string]any{"name": "Text Test Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	botID := created["id"].(string)

	// Create text source
	var mbody bytes.Buffer
	mw := multipart.NewWriter(&mbody)
	mw.WriteField("source_type", "text")
	mw.WriteField("text", "This is test content for the source")
	mw.Close()

	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody.Bytes()))
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusCreated {
		t.Fatalf("create text source: got %d, want %d, body: %s", rr2.Code, http.StatusCreated, rr2.Body.String())
	}

	var resp map[string]string
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp)
	if resp["id"] == "" {
		t.Fatal("source id should not be empty")
	}
}

// TestSources_URLCreation tests URL source creation end-to-end
func TestSources_URLCreation(t *testing.T) {
	dbx, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Skipf("db not available: %v", err)
	}
	defer dbx.Close()

	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("url_create+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	ch := &ChatbotHandlers{DB: dbx}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage()}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	// Create chatbot
	cb := map[string]any{"name": "URL Test Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	botID := created["id"].(string)

	// Create URL source
	var mbody bytes.Buffer
	mw := multipart.NewWriter(&mbody)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", "https://example.com/test-"+fmt.Sprint(time.Now().UnixNano()))
	mw.Close()

	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody.Bytes()))
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusCreated {
		t.Fatalf("create URL source: got %d, want %d, body: %s", rr2.Code, http.StatusCreated, rr2.Body.String())
	}

	var resp map[string]string
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp)
	if resp["id"] == "" {
		t.Fatal("source id should not be empty")
	}
}

// TestSources_EmptyText_BadRequest tests empty text rejection
func TestSources_EmptyText_BadRequest(t *testing.T) {
	dbx, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Skipf("db not available: %v", err)
	}
	defer dbx.Close()

	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("empty_text+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	ch := &ChatbotHandlers{DB: dbx}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage()}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	// Create chatbot
	cb := map[string]any{"name": "Empty Text Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	botID := created["id"].(string)

	// Try to create text source with empty text
	var mbody bytes.Buffer
	mw := multipart.NewWriter(&mbody)
	mw.WriteField("source_type", "text")
	mw.WriteField("text", "   ") // Empty/whitespace text
	mw.Close()

	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody.Bytes()))
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("empty text source: got %d, want %d", rr2.Code, http.StatusBadRequest)
	}
}

// TestSources_EmptyURL_BadRequest tests empty URL rejection
func TestSources_EmptyURL_BadRequest(t *testing.T) {
	dbx, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Skipf("db not available: %v", err)
	}
	defer dbx.Close()

	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("empty_url+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	ch := &ChatbotHandlers{DB: dbx}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage()}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	// Create chatbot
	cb := map[string]any{"name": "Empty URL Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	botID := created["id"].(string)

	// Try to create URL source with empty URL
	var mbody bytes.Buffer
	mw := multipart.NewWriter(&mbody)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", "   ") // Empty/whitespace URL
	mw.Close()

	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody.Bytes()))
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("empty URL source: got %d, want %d", rr2.Code, http.StatusBadRequest)
	}
}

// TestSources_DuplicateURL_Conflict tests duplicate URL rejection
func TestSources_DuplicateURL_Conflict(t *testing.T) {
	dbx, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Skipf("db not available: %v", err)
	}
	defer dbx.Close()

	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("dup_url+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	ch := &ChatbotHandlers{DB: dbx}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage()}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	// Create chatbot
	cb := map[string]any{"name": "Dup URL Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	botID := created["id"].(string)

	testURL := "https://example.com/unique-" + fmt.Sprint(time.Now().UnixNano())

	// Create first URL source
	var mbody1 bytes.Buffer
	mw1 := multipart.NewWriter(&mbody1)
	mw1.WriteField("source_type", "url")
	mw1.WriteField("source_url", testURL)
	mw1.Close()

	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody1.Bytes()))
	r2.Header.Set("Content-Type", mw1.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusCreated {
		t.Fatalf("first URL source: got %d, want %d", rr2.Code, http.StatusCreated)
	}

	// Try to create duplicate URL source
	var mbody2 bytes.Buffer
	mw2 := multipart.NewWriter(&mbody2)
	mw2.WriteField("source_type", "url")
	mw2.WriteField("source_url", testURL)
	mw2.Close()

	r3 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody2.Bytes()))
	r3.Header.Set("Content-Type", mw2.FormDataContentType())
	rr3 := httptest.NewRecorder()
	sh.ChatbotSources(rr3, ctx(r3))

	if rr3.Code != http.StatusConflict && rr3.Code != http.StatusForbidden {
		t.Fatalf("duplicate URL source: got %d, want %d or %d", rr3.Code, http.StatusConflict, http.StatusForbidden)
	}
}

// TestSources_InvalidSourceType_BadRequest tests invalid source type
func TestSources_InvalidSourceType_BadRequest(t *testing.T) {
	dbx, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Skipf("db not available: %v", err)
	}
	defer dbx.Close()

	var uid string
	var freePlanID string
	if err := dbx.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("invalid_type+%d@example.com", time.Now().UnixNano())
	if err := dbx.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "x", freePlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	ch := &ChatbotHandlers{DB: dbx}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage()}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	// Create chatbot
	cb := map[string]any{"name": "Invalid Type Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	botID := created["id"].(string)

	// Try to create source with invalid type
	var mbody bytes.Buffer
	mw := multipart.NewWriter(&mbody)
	mw.WriteField("source_type", "invalid_type")
	mw.Close()

	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody.Bytes()))
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("invalid source type: got %d, want %d", rr2.Code, http.StatusBadRequest)
	}
}
