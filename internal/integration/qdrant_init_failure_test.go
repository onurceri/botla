package integration

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func startQdrantInitFailStub() *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/collections/embeddings", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	h.HandleFunc("/collections/embeddings/points", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	h.HandleFunc("/collections/embeddings/points/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": []any{}})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	h.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	return httptest.NewServer(h)
}

func TestStartup_QdrantCollectionInitFailure_StillWorks(t *testing.T) {
	oai := startOpenAIStub()
	qd := startQdrantInitFailStub()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	token := authToken(t, te.Server.URL, "qdinitfail@example.com")
	create := map[string]any{"name": "QD Init Fail Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	var b strings.Builder
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "text")
	mw.WriteField("text", "init failure test")
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
	sourceID := sid["id"]
	if sourceID == "" {
		t.Fatalf("source id empty")
	}

	statusPath := "/api/v1/sources/" + url.PathEscape(sourceID)
	completed := false
	for i := 0; i < 200; i++ {
		reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+statusPath, nil)
		reqG.Header.Set("Authorization", "Bearer "+token)
		resG, _ := http.DefaultClient.Do(reqG)
		if resG.StatusCode != http.StatusOK {
			resG.Body.Close()
			continue
		}
		var st map[string]any
		json.NewDecoder(resG.Body).Decode(&st)
		resG.Body.Close()
		s := st["status"].(string)
		if s == "completed" {
			completed = true
			break
		}
	}
	if !completed {
		t.Fatalf("expected completed despite init failure")
	}
}
