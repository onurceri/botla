package integration

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
)

func startLinkedHTMLStub() *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Root</h1><a href='/page1'>Page 1</a></body></html>"))
	})
	h.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Page 1</h1><a href='/page2'>Page 2</a></body></html>"))
	})
	h.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Page 2</h1><p>End</p></body></html>"))
	})
	return httptest.NewServer(h)
}

func TestURLDiscovery_AutoMode(t *testing.T) {
	oai := startOpenAIStub()
	qd := startQdrantStub()
	page := startLinkedHTMLStub()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()
	defer page.Close()

	token := authToken(t, te.Server.URL, "autodisc@example.com")

	// Create chatbot with auto discovery
	create := map[string]any{
		"name":           "Auto Discovery Bot",
		"discovery_mode": "auto",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Add root source
	var b strings.Builder
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", page.URL)
	mw.Close()
	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(b.String()))
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := http.DefaultClient.Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	rootSourceID := sid["id"]

	// Wait for processing to complete
	waitForProcessing(t, te, token, rootSourceID)

	// Check if Page 1 was discovered and added as a source
	// We expect 2 sources: Root and Page 1
	// Page 2 should NOT be added because Page 1 is "is_discovered=true" and recursion stops there.

	time.Sleep(500 * time.Millisecond) // Give a little extra time for the async discovery to trigger and finish

	sources, err := db.ListSourcesByChatbotID(context.Background(), te.DB, bot.ID)
	if err != nil {
		t.Fatalf("failed to list sources: %v", err)
	}

	if len(sources) != 2 {
		t.Errorf("expected 2 sources (root + page1), got %d", len(sources))
		for _, s := range sources {
			t.Logf("Source: %s (Discovered: %v)", *s.SourceURL, s.IsDiscovered)
		}
	}

	foundPage1 := false
	for _, s := range sources {
		if strings.Contains(*s.SourceURL, "/page1") {
			foundPage1 = true
			if !s.IsDiscovered {
				t.Error("Page 1 should be marked as discovered")
			}
		} else if strings.Contains(*s.SourceURL, "/page2") {
			t.Error("Page 2 should NOT be discovered (recursion limit)")
		}
	}

	if !foundPage1 {
		t.Error("Page 1 was not found in sources")
	}
}

func TestURLDiscovery_PendingMode(t *testing.T) {
	oai := startOpenAIStub()
	qd := startQdrantStub()
	page := startLinkedHTMLStub()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()
	defer page.Close()

	token := authToken(t, te.Server.URL, "pendingdisc@example.com")

	// Create chatbot with pending discovery
	create := map[string]any{
		"name":           "Pending Discovery Bot",
		"discovery_mode": "pending",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Add root source
	var b strings.Builder
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", page.URL)
	mw.Close()
	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(b.String()))
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := http.DefaultClient.Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	rootSourceID := sid["id"]

	// Wait for processing to complete
	waitForProcessing(t, te, token, rootSourceID)

	time.Sleep(500 * time.Millisecond)

	// Check that Page 1 is NOT in sources
	sources, err := db.ListSourcesByChatbotID(context.Background(), te.DB, bot.ID)
	if err != nil {
		t.Fatalf("failed to list sources: %v", err)
	}
	if len(sources) != 1 {
		t.Errorf("expected only 1 source (root), got %d", len(sources))
	}

	// Check that Page 1 IS in pending_discovered_urls
	// We need to query the pending_discovered_urls table.
	// Since there is no public API to list pending URLs (or I assume so), I'll use DB directly.

	rows, err := te.DB.Query("SELECT id, url FROM pending_discovered_urls WHERE chatbot_id = $1", bot.ID)
	if err != nil {
		t.Fatalf("failed to query pending_discovered_urls: %v", err)
	}
	defer rows.Close()

	var pendingURLs []struct {
		ID  string
		URL string
	}
	for rows.Next() {
		var u struct {
			ID  string
			URL string
		}
		if scanErr := rows.Scan(&u.ID, &u.URL); scanErr != nil {
			t.Fatalf("failed to scan pending url: %v", scanErr)
		}
		pendingURLs = append(pendingURLs, u)
	}

	foundPage1 := false
	var page1ID string
	for _, u := range pendingURLs {
		if strings.Contains(u.URL, "/page1") {
			foundPage1 = true
			page1ID = u.ID
		}
	}

	if !foundPage1 {
		t.Errorf("Page 1 not found in pending_discovered_urls. Found: %v", pendingURLs)
	}

	// Now approve Page 1
	approveReq := map[string]any{
		"url_ids": []string{page1ID},
	}
	approveBody, _ := json.Marshal(approveReq)
	reqA, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/pending-urls/approve", strings.NewReader(string(approveBody)))
	reqA.Header.Set("Authorization", "Bearer "+token)
	reqA.Header.Set("Content-Type", "application/json")
	resA, _ := http.DefaultClient.Do(reqA)
	if resA.StatusCode != http.StatusOK {
		t.Errorf("approve failed, status: %d", resA.StatusCode)
	}
	resA.Body.Close()

	// Wait for the new source to be processed (best effort wait)
	time.Sleep(1 * time.Second)

	// Verify Page 1 is now a source
	sources, err = db.ListSourcesByChatbotID(context.Background(), te.DB, bot.ID)
	if err != nil {
		t.Fatalf("failed to list sources: %v", err)
	}
	foundPage1Source := false
	for _, s := range sources {
		if strings.Contains(*s.SourceURL, "/page1") {
			foundPage1Source = true
			if !s.IsDiscovered {
				t.Error("Page 1 should be marked as discovered")
			}
		}
	}
	if !foundPage1Source {
		t.Error("Page 1 was NOT promoted to source after approval")
	}

	// Verify Page 1 status is updated in pending_discovered_urls
	var status string
	err = te.DB.QueryRow("SELECT status FROM pending_discovered_urls WHERE id = $1", page1ID).Scan(&status)
	if err != nil {
		t.Fatalf("failed to query pending url status: %v", err)
	}
	if status != "selected" { // Assuming 'selected' or similar is used when approved. Checking handler code, it creates source but doesn't explicitly delete from pending?
		// Wait, looking at handler code:
		// db.CreateDiscoveredSource is called.
		// What happens to the pending entry?
		// I need to check PendingURLsHandlers.ApprovePendingURLs logic again.
		// It calls db.UpdatePendingURLStatus(ctx, h.DB, chatbotID, req.IDs, "selected")
		// So status should be 'selected'.
		t.Errorf("expected pending url status 'selected', got '%s'", status)
	}
}

func waitForProcessing(t *testing.T, te *TestEnv, token, sourceID string) {
	statusPath := "/api/v1/sources/" + url.PathEscape(sourceID)
	for i := 0; i < 200; i++ {
		reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+statusPath, nil)
		reqG.Header.Set("Authorization", "Bearer "+token)
		resG, _ := http.DefaultClient.Do(reqG)
		if resG.StatusCode != http.StatusOK {
			resG.Body.Close()
			time.Sleep(20 * time.Millisecond)
			continue
		}
		var st map[string]any
		json.NewDecoder(resG.Body).Decode(&st)
		resG.Body.Close()
		s := st["status"].(string)
		if s != "pending" {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("source processing timed out")
}
