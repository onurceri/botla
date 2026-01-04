package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/pkg/config"
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

func TestEmbeddingService_Generate_RetryAndWarn(t *testing.T) {
	t.Parallel()
	of := true
	qf := true
	oai := newOpenAIServerBatch(&of)
	defer oai.Close()
	qdr := newQdrantServerUpsert(&qf)
	defer qdr.Close()
	chunks := []models.Chunk{{Text: "hello", TokenCount: 2}}
	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: oai.URL,
	}
	emb, _ := NewOpenAIClient(cfg)
	vc, _ := NewQdrantClient(&QdrantConfig{URL: qdr.URL})
	svc := NewEmbeddingService(emb, vc, nil)
	if err := svc.Generate(context.Background(), chunks, "cb"); err != nil {
		t.Fatalf("gen err: %v", err)
	}
}

func TestEmbeddingService_GenerateForSource_UpsertError(t *testing.T) {
	t.Parallel()
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
	chunks := []models.Chunk{{Text: "hello", TokenCount: 2}}
	cfg := &config.Config{
		OPENAI_API_KEY:  "k",
		OPENAI_API_BASE: oai.URL,
	}
	emb, _ := NewOpenAIClient(cfg)
	vc, _ := NewQdrantClient(&QdrantConfig{URL: qdr.URL})
	svc := NewEmbeddingService(emb, vc, nil)
	if err := svc.GenerateForSource(context.Background(), chunks, "cb", "src", "file"); err == nil {
		t.Fatalf("expected error")
	}
}

// EMB-001: Generate embeddings for 0 chunks
func TestEmbeddingService_Generate_Empty(t *testing.T) {
	svc := NewEmbeddingService(nil, nil, nil)
	err := svc.Generate(context.Background(), nil, "cb")
	if err != nil {
		t.Fatalf("expected nil error for empty input, got %v", err)
	}
	err = svc.Generate(context.Background(), []models.Chunk{}, "cb")
	if err != nil {
		t.Fatalf("expected nil error for empty chunks, got %v", err)
	}
}

// EMB-003: Generate embeddings for 26 chunks (batching)
func TestEmbeddingService_Generate_Batching(t *testing.T) {
	t.Parallel()
	// Mock OpenAI to count requests
	reqCount := 0
	oaiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount++
		// Decode body to check batch size if we want, but just counting requests is enough
		// We expect 2 requests for 26 chunks (25 + 1)

		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		inputs, _ := req["input"].([]any)

		respEmbeddings := make([]map[string]any, len(inputs))
		for i := range inputs {
			respEmbeddings[i] = map[string]any{"embedding": []float64{0.1}}
		}

		json.NewEncoder(w).Encode(map[string]any{
			"data":  respEmbeddings,
			"usage": map[string]int{"total_tokens": 10},
		})
	}))
	defer oaiSrv.Close()

	// Mock Qdrant to accept upserts
	qdrSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer qdrSrv.Close()

	chunks := make([]models.Chunk, 26)
	for i := 0; i < 26; i++ {
		chunks[i] = models.Chunk{Text: "a", TokenCount: 1}
	}

	cfg := &config.Config{
		OPENAI_API_KEY:  "dummy",
		OPENAI_API_BASE: oaiSrv.URL,
	}
	emb, _ := NewOpenAIClient(cfg)
	vc, _ := NewQdrantClient(&QdrantConfig{URL: qdrSrv.URL})

	// We need to speed up the ticker or wait.
	// The code uses time.NewTicker(time.Second / 58) which is ~17ms.
	// 2 batches = ~34ms wait. This is fast enough for a test.

	svc := NewEmbeddingService(emb, vc, nil)
	err := svc.Generate(context.Background(), chunks, "cb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reqCount != 2 {
		t.Errorf("expected 2 OpenAI requests (batches), got %d", reqCount)
	}
}
