package integration

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
)

func startQdrantStub() *httptest.Server {
	h := http.NewServeMux()

	// Handle any collection - supports dynamic collection names like embeddings_it_xxx
	h.HandleFunc("/collections/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		w.Header().Set("Content-Type", "application/json")

		// GET/PUT collection (e.g. /collections/embeddings_it_xxx)
		if !strings.Contains(path, "/points") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		// Points operations
		switch {
		case strings.HasSuffix(path, "/points"):
			// PUT points - upsert
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
				return
			}
		case strings.HasSuffix(path, "/points/delete"):
			// POST delete
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
				return
			}
		case strings.HasSuffix(path, "/points/search"):
			// POST search
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				result := []map[string]any{{
					"id":    "p1",
					"score": 0.9,
					"payload": map[string]any{
						"chatbot_id":    "stub-bot",
						"source_id":     "00000000-0000-0000-0000-000000000001",
						"chunk_index":   0,
						"original_text": "stub text chunk",
						"source_type":   "text",
					},
				}}
				json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": result})
				return
			}
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	h.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	return httptest.NewServer(h)
}

type QdrantTopKStub struct {
	Server          *httptest.Server
	Mu              sync.Mutex
	LastSearchLimit int
	SearchCalls     int
}

func (s *QdrantTopKStub) recordSearchLimit(limit int) {
	s.Mu.Lock()
	s.LastSearchLimit = limit
	s.SearchCalls++
	s.Mu.Unlock()
}

func startQdrantTopKStub() *QdrantTopKStub {
	s := &QdrantTopKStub{}
	h := http.NewServeMux()

	// Handle all collection operations with dynamic names
	h.HandleFunc("/collections/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		w.Header().Set("Content-Type", "application/json")

		// Collection creation/info (e.g. PUT /collections/embeddings_xxx)
		if !strings.Contains(path, "/points") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "ok",
				"result": map[string]any{
					"status": "green",
				},
			})
			return
		}

		// POST search
		if strings.HasSuffix(path, "/points/search") && r.Method == http.MethodPost {
			var reqBody struct {
				Limit int `json:"limit"`
			}
			_ = json.NewDecoder(r.Body).Decode(&reqBody)
			s.recordSearchLimit(reqBody.Limit)

			w.WriteHeader(http.StatusOK)
			result := []map[string]any{{
				"id":    "p1",
				"score": 0.9,
				"payload": map[string]any{
					"chatbot_id":    "00000000-0000-0000-0000-000000000001",
					"source_id":     "00000000-0000-0000-0000-000000000001",
					"chunk_index":   0,
					"original_text": "stub text chunk",
					"source_type":   "text",
				},
			}}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "result": result})
			return
		}

		// PUT points - upsert
		if strings.HasSuffix(path, "/points") && r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		// POST delete
		if strings.HasSuffix(path, "/points/delete") && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Health endpoint
	h.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s.Server = httptest.NewServer(h)
	return s
}

func TestSources_Text_Ingest_Status_Delete(t *testing.T) {
	t.Parallel()
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	defer oai.Close()
	defer qd.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "sources@example.com")

	// create chatbot
	create := map[string]any{"name": "Src Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// upload text source
	var b strings.Builder
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "text")
	mw.WriteField("text", "bu bir test içeriğidir")
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
	if sourceID == "" {
		t.Fatalf("source id empty")
	}

	// poll status
	statusPath := "/api/v1/sources/" + url.PathEscape(sourceID)
	for i := 0; i < 40; i++ {
		reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+statusPath, nil)
		reqG.Header.Set("Authorization", "Bearer "+token)
		resG, _ := testHTTPClient().Do(reqG)
		if resG.StatusCode != http.StatusOK {
			resG.Body.Close()
			continue
		}
		var st map[string]any
		json.NewDecoder(resG.Body).Decode(&st)
		resG.Body.Close()
		s := st["status"].(string)
		if s == "completed" {
			cc := int(st["chunk_count"].(float64))
			if cc < 1 {
				t.Fatalf("expected chunk_count>0")
			}
			break
		}
	}

	// delete
	reqD, _ := http.NewRequest(http.MethodDelete, te.Server.URL+statusPath, nil)
	reqD.Header.Set("Authorization", "Bearer "+token)
	resD, _ := testHTTPClient().Do(reqD)
	if resD.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resD.StatusCode)
	}
	resD.Body.Close()
}
