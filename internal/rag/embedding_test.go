package rag

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
)

// simulate OpenAI embeddings with first failure then success
func newOpenAIServerBatch(firstFail *bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/embeddings" {
			if *firstFail {
				*firstFail = false
				http.Error(w, "oops", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data":  []map[string]any{{"embedding": []float64{0.1, 0.2}}},
				"usage": map[string]int{"prompt_tokens": 1, "total_tokens": 2},
			})
			return
		}
		http.NotFound(w, r)
	}))
}

// qdrant server with first upsert failure then success
func newQdrantServerUpsert(firstFail *bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/collections/embeddings/points" {
			if *firstFail {
				*firstFail = false
				http.Error(w, "fail", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
}

func TestGenerateEmbeddings_RetryAndWarn(t *testing.T) {
	of := true
	qf := true
	oai := newOpenAIServerBatch(&of)
	defer oai.Close()
	qdr := newQdrantServerUpsert(&qf)
	defer qdr.Close()
	t.Setenv("OPENAI_API_KEY", "k")
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qdr.URL)
	chunks := []models.Chunk{{Text: "hello", TokenCount: 2}}
	if err := GenerateEmbeddings(chunks, "cb"); err != nil {
		t.Fatalf("gen err: %v", err)
	}
}

func TestGenerateEmbeddingsForSource_UpsertError(t *testing.T) {
	of := false
	oai := newOpenAIServerBatch(&of)
	defer oai.Close()
	// qdrant always fail upsert
	qdr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/collections/embeddings/points" {
			http.Error(w, "fail", http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer qdr.Close()
	t.Setenv("OPENAI_API_KEY", "k")
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qdr.URL)
	chunks := []models.Chunk{{Text: "hello", TokenCount: 2}}
	if err := GenerateEmbeddingsForSource(chunks, "cb", "src", "file"); err == nil {
		t.Fatalf("expected error")
	}
}
