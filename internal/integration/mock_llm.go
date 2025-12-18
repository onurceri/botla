package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// LLMMock is a configurable mock server for LLM providers (OpenAI compatible)
type LLMMock struct {
	Server          *httptest.Server
	URL             string
	Requests        []MockRequest
	mu              sync.Mutex
	ChatResponse    func(req MockRequest) (map[string]any, int)
	EmbedResponse   func(req MockRequest) (map[string]any, int)
	DefaultResponse string
}

type MockRequest struct {
	Method string
	Path   string
	Body   map[string]any
}

// NewLLMMock creates a new LLM mock server
func NewLLMMock(t *testing.T) *LLMMock {
	m := &LLMMock{
		DefaultResponse: "This is a mocked response.",
	}

	mux := http.NewServeMux()

	// Chat Completions Handler
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]any
		_ = json.Unmarshal(body, &reqBody)

		captured := MockRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   reqBody,
		}

		m.mu.Lock()
		m.Requests = append(m.Requests, captured)
		customHandler := m.ChatResponse
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		if customHandler != nil {
			resp, status := customHandler(captured)
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Default behavior
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":      "chatcmpl-mock",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   "gpt-4o-mock",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": m.DefaultResponse,
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 10,
				"total_tokens":      20,
			},
		})
	})

	// Embeddings Handler
	mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]any
		_ = json.Unmarshal(body, &reqBody)

		captured := MockRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   reqBody,
		}

		m.mu.Lock()
		m.Requests = append(m.Requests, captured)
		customHandler := m.EmbedResponse
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		if customHandler != nil {
			resp, status := customHandler(captured)
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Default behavior: return 1536-dim vector
		input := reqBody["input"]
		var count int
		switch v := input.(type) {
		case string:
			count = 1
		case []interface{}:
			count = len(v)
		default:
			count = 1
		}

		data := make([]map[string]any, count)
		for i := 0; i < count; i++ {
			embedding := make([]float64, 1536)
			for j := range embedding {
				embedding[j] = 0.001
			}
			data[i] = map[string]any{
				"object":    "embedding",
				"index":     i,
				"embedding": embedding,
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data":   data,
			"model":  "text-embedding-3-small",
			"usage": map[string]int{
				"prompt_tokens": 10,
				"total_tokens":  10,
			},
		})
	})

	m.Server = httptest.NewServer(mux)
	m.URL = m.Server.URL
	return m
}

func (m *LLMMock) Close() {
	m.Server.Close()
}

func (m *LLMMock) SetChatResponse(handler func(req MockRequest) (map[string]any, int)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ChatResponse = handler
}

func (m *LLMMock) SetEmbedResponse(handler func(req MockRequest) (map[string]any, int)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EmbedResponse = handler
}
