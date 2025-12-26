package qdrant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/ai"
)

func TestClient_ImplementsVectorStore(t *testing.T) {
	var _ ai.VectorStore = (*Client)(nil)
}

func TestNew(t *testing.T) {
	client := New("http://localhost:6333", "test-key", nil)
	if client == nil {
		t.Fatal("expected client to be created")
	}
	if client.baseURL != "http://localhost:6333" {
		t.Errorf("expected baseURL to be 'http://localhost:6333', got %s", client.baseURL)
	}
	if client.apiKey != "test-key" {
		t.Errorf("expected apiKey to be 'test-key', got %s", client.apiKey)
	}
	if client.http == nil {
		t.Error("expected http client to be set")
	}
}

func TestNewFromEnv_MissingURL(t *testing.T) {
	t.Setenv("QDRANT_URL", "")
	t.Setenv("QDRANT_API_KEY", "test-key")

	_, err := NewFromEnv()
	if err == nil {
		t.Error("expected error when QDRANT_URL is empty")
	}
}

func TestNewFromEnv_Success(t *testing.T) {
	t.Setenv("QDRANT_URL", "http://localhost:6333")
	t.Setenv("QDRANT_API_KEY", "test-key")
	t.Setenv("QDRANT_TIMEOUT_MS", "5000")

	client, err := NewFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL != "http://localhost:6333" {
		t.Errorf("expected baseURL to be 'http://localhost:6333', got %s", client.baseURL)
	}
	if client.apiKey != "test-key" {
		t.Errorf("expected apiKey to be 'test-key', got %s", client.apiKey)
	}
}

func TestEnsureCollection_AlreadyExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := New(server.URL, "test-key", server.Client())
	ctx := context.Background()

	err := client.EnsureCollection(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEnsureCollection_Create(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path == "/collections/embeddings" {
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
				return
			}
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := New(server.URL, "test-key", server.Client())
	ctx := context.Background()

	err := client.EnsureCollection(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if requestCount != 2 {
		t.Errorf("expected 2 requests (GET + PUT), got %d", requestCount)
	}
}

func TestUpsert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points" && r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := New(server.URL, "test-key", server.Client())
	ctx := context.Background()

	payload := ai.VectorPayload{
		ChatbotID:    "bot-123",
		SourceID:     "src-456",
		ChunkIndex:   0,
		OriginalText: "test text",
		SourceType:   "text",
		CreatedAt:    time.Now(),
	}

	err := client.Upsert(ctx, "vec-1", []float32{0.1, 0.2, 0.3}, payload)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points/search" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"status":"ok",
				"result":[
					{
						"id":"vec-1",
						"score":0.95,
						"payload":{
							"chatbot_id":"bot-123",
							"source_id":"src-456",
							"chunk_index":0,
							"original_text":"test text",
							"source_type":"text",
							"created_at":"2024-01-01T00:00:00Z"
						}
					}
				]
			}`))
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := New(server.URL, "test-key", server.Client())
	ctx := context.Background()

	filter := ai.SearchFilter{ChatbotID: "bot-123"}
	results, err := client.Search(ctx, []float32{0.1, 0.2, 0.3}, filter, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].Score != 0.95 {
		t.Errorf("expected score 0.95, got %f", results[0].Score)
	}
}

func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points/delete" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := New(server.URL, "test-key", server.Client())
	ctx := context.Background()

	filter := ai.DeleteFilter{SourceID: "src-456"}
	err := client.Delete(ctx, filter)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestScroll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points/scroll" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"status":"ok",
				"result":{
					"points":[
						{
							"id":"vec-1",
							"score":0.95,
							"payload":{
								"chatbot_id":"bot-123",
								"source_id":"src-456",
								"chunk_index":0,
								"original_text":"test text",
								"source_type":"text",
								"created_at":"2024-01-01T00:00:00Z"
							}
						}
					],
					"next_page_offset":"next-offset"
				}
			}`))
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := New(server.URL, "test-key", server.Client())
	ctx := context.Background()

	filter := ai.SearchFilter{SourceID: "src-456"}
	results, nextOffset, err := client.Scroll(ctx, filter, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if nextOffset == nil {
		t.Error("expected nextOffset to be set")
	}
}
