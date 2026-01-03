package integration

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
)

func startHTMLStub() *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Stub Page</h1><p>İçerik testi</p></body></html>"))
	})
	return httptest.NewServer(h)
}

func TestSources_URL_Ingest_Success(t *testing.T) {
	t.Parallel()
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	page := startHTMLStub()
	defer oai.Close()
	defer qd.Close()
	defer page.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "urlsrc@example.com")
	create := map[string]any{"name": "URL Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	var b strings.Builder
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "url")
	mw.WriteField("source_url", page.URL)
	mw.Close()
	reqS, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(b.String()))
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := testHTTPClient().Do(reqS)
	if resS.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resS.StatusCode)
	}
	var sid map[string]string
	json.NewDecoder(resS.Body).Decode(&sid)
	resS.Body.Close()
	sourceID := sid["id"]

	statusPath := "/api/v1/sources/" + url.PathEscape(sourceID)
	notPending := false
	for i := 0; i < 200; i++ {
		reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+statusPath, nil)
		reqG.Header.Set("Authorization", "Bearer "+token)
		resG, _ := testHTTPClient().Do(reqG)
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
			notPending = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !notPending {
		t.Fatalf("status still pending")
	}
}
