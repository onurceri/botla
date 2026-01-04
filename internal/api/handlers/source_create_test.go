package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/middleware"
	"github.com/onurceri/botla-app/pkg/storage"
)

// TestSources_TextCreation tests text source creation end-to-end
func TestSources_TextCreation(t *testing.T) {
	dbx := testdb.OpenTestDB(t)

	user := testdb.CreateUser(t, dbx)
	uid := user.ID

	chatbotRepo := repository.NewPostgresChatbotRepo(dbx)
	planRepo := repository.NewPostgresPlanRepo(dbx, nil)
	ch := &ChatbotHandlers{DB: dbx, ChatbotRepo: chatbotRepo, PlanRepo: planRepo}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage(), SourceRepo: repository.NewPostgresSourceRepo(dbx), ChatbotRepo: chatbotRepo, UsageRepo: repository.NewPostgresUsageRepo(dbx), PlanRepo: planRepo}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	// Create chatbot
	cb := map[string]any{"name": "Text Test Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	if rr1.Code != http.StatusCreated {
		t.Fatalf("create chatbot: %d, body: %s", rr1.Code, rr1.Body.String())
	}
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
	r2.SetPathValue("id", botID)
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
	dbx := testdb.OpenTestDB(t)

	user := testdb.CreateUser(t, dbx)
	uid := user.ID

	chatbotRepo := repository.NewPostgresChatbotRepo(dbx)
	planRepo := repository.NewPostgresPlanRepo(dbx, nil)
	ch := &ChatbotHandlers{DB: dbx, ChatbotRepo: chatbotRepo, PlanRepo: planRepo}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage(), SourceRepo: repository.NewPostgresSourceRepo(dbx), ChatbotRepo: chatbotRepo, UsageRepo: repository.NewPostgresUsageRepo(dbx), PlanRepo: planRepo}
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
	r2.SetPathValue("id", botID)
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
	dbx := testdb.OpenTestDB(t)

	user := testdb.CreateUser(t, dbx)
	uid := user.ID

	chatbotRepo := repository.NewPostgresChatbotRepo(dbx)
	planRepo := repository.NewPostgresPlanRepo(dbx, nil)
	ch := &ChatbotHandlers{DB: dbx, ChatbotRepo: chatbotRepo, PlanRepo: planRepo}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage(), SourceRepo: repository.NewPostgresSourceRepo(dbx), ChatbotRepo: chatbotRepo, UsageRepo: repository.NewPostgresUsageRepo(dbx), PlanRepo: planRepo}
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
	r2.SetPathValue("id", botID)
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("empty text source: got %d, want %d", rr2.Code, http.StatusBadRequest)
	}
}

// TestSources_EmptyURL_BadRequest tests empty URL rejection
func TestSources_EmptyURL_BadRequest(t *testing.T) {
	dbx := testdb.OpenTestDB(t)

	user := testdb.CreateUser(t, dbx)
	uid := user.ID

	chatbotRepo := repository.NewPostgresChatbotRepo(dbx)
	planRepo := repository.NewPostgresPlanRepo(dbx, nil)
	ch := &ChatbotHandlers{DB: dbx, ChatbotRepo: chatbotRepo, PlanRepo: planRepo}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage(), SourceRepo: repository.NewPostgresSourceRepo(dbx), ChatbotRepo: chatbotRepo, UsageRepo: repository.NewPostgresUsageRepo(dbx), PlanRepo: planRepo}
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
	r2.SetPathValue("id", botID)
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("empty URL source: got %d, want %d", rr2.Code, http.StatusBadRequest)
	}
}

// TestSources_DuplicateURL_Conflict tests duplicate URL rejection
func TestSources_DuplicateURL_Conflict(t *testing.T) {
	dbx := testdb.OpenTestDB(t)

	user := testdb.CreateUser(t, dbx)
	uid := user.ID

	chatbotRepo := repository.NewPostgresChatbotRepo(dbx)
	planRepo := repository.NewPostgresPlanRepo(dbx, nil)
	ch := &ChatbotHandlers{DB: dbx, ChatbotRepo: chatbotRepo, PlanRepo: planRepo}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage(), SourceRepo: repository.NewPostgresSourceRepo(dbx), ChatbotRepo: chatbotRepo, UsageRepo: repository.NewPostgresUsageRepo(dbx), PlanRepo: planRepo}
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
	r2.SetPathValue("id", botID)
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
	r3.SetPathValue("id", botID)
	r3.Header.Set("Content-Type", mw2.FormDataContentType())
	rr3 := httptest.NewRecorder()
	sh.ChatbotSources(rr3, ctx(r3))

	if rr3.Code != http.StatusConflict && rr3.Code != http.StatusForbidden {
		t.Fatalf("duplicate URL source: got %d, want %d or %d", rr3.Code, http.StatusConflict, http.StatusForbidden)
	}
}

// TestSources_InvalidSourceType_BadRequest tests invalid source type
func TestSources_InvalidSourceType_BadRequest(t *testing.T) {
	dbx := testdb.OpenTestDB(t)

	user := testdb.CreateUser(t, dbx)
	uid := user.ID

	chatbotRepo := repository.NewPostgresChatbotRepo(dbx)
	planRepo := repository.NewPostgresPlanRepo(dbx, nil)
	ch := &ChatbotHandlers{DB: dbx, ChatbotRepo: chatbotRepo, PlanRepo: planRepo}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage(), SourceRepo: repository.NewPostgresSourceRepo(dbx), ChatbotRepo: chatbotRepo, UsageRepo: repository.NewPostgresUsageRepo(dbx), PlanRepo: planRepo}
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
	r2.SetPathValue("id", botID)
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("invalid source type: got %d, want %d", rr2.Code, http.StatusBadRequest)
	}
}

// TestSources_DuplicateURL_TrailingSlash_Conflict tests that URLs with/without trailing slash are detected as duplicates
func TestSources_DuplicateURL_TrailingSlash_Conflict(t *testing.T) {
	dbx := testdb.OpenTestDB(t)

	user := testdb.CreateUser(t, dbx)
	uid := user.ID

	chatbotRepo := repository.NewPostgresChatbotRepo(dbx)
	planRepo := repository.NewPostgresPlanRepo(dbx, nil)
	ch := &ChatbotHandlers{DB: dbx, ChatbotRepo: chatbotRepo, PlanRepo: planRepo}
	sh := &SourcesHandlers{DB: dbx, Storage: storage.NewMemoryStorage(), SourceRepo: repository.NewPostgresSourceRepo(dbx), ChatbotRepo: chatbotRepo, UsageRepo: repository.NewPostgresUsageRepo(dbx), PlanRepo: planRepo}
	ctx := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uid))
	}

	// Create chatbot
	cb := map[string]any{"name": "Trailing Slash Test Bot"}
	jb, _ := json.Marshal(cb)
	r1 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader(jb))
	rr1 := httptest.NewRecorder()
	ch.ListOrCreate(rr1, ctx(r1))
	var created map[string]any
	_ = json.Unmarshal(rr1.Body.Bytes(), &created)
	botID := created["id"].(string)

	baseURL := "https://espacefarmento.fr"

	// Create first URL source WITH trailing slash
	var mbody1 bytes.Buffer
	mw1 := multipart.NewWriter(&mbody1)
	mw1.WriteField("source_type", "url")
	mw1.WriteField("source_url", baseURL+"/") // With trailing slash
	mw1.Close()

	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody1.Bytes()))
	r2.SetPathValue("id", botID)
	r2.Header.Set("Content-Type", mw1.FormDataContentType())
	rr2 := httptest.NewRecorder()
	sh.ChatbotSources(rr2, ctx(r2))

	if rr2.Code != http.StatusCreated {
		t.Fatalf("first URL source (with slash): got %d, want %d, body: %s", rr2.Code, http.StatusCreated, rr2.Body.String())
	}

	// Try to create same URL source WITHOUT trailing slash - should be detected as duplicate
	var mbody2 bytes.Buffer
	mw2 := multipart.NewWriter(&mbody2)
	mw2.WriteField("source_type", "url")
	mw2.WriteField("source_url", baseURL) // Without trailing slash
	mw2.Close()

	r3 := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+botID+"/sources", bytes.NewReader(mbody2.Bytes()))
	r3.SetPathValue("id", botID)
	r3.Header.Set("Content-Type", mw2.FormDataContentType())
	rr3 := httptest.NewRecorder()
	sh.ChatbotSources(rr3, ctx(r3))

	// Should be detected as duplicate due to URL normalization
	if rr3.Code != http.StatusConflict && rr3.Code != http.StatusForbidden {
		t.Fatalf("duplicate URL (without slash): got %d, want %d or %d - URL normalization may not be working", rr3.Code, http.StatusConflict, http.StatusForbidden)
	}
}
