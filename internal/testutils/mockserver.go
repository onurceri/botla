package testutils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// MockServer provides an HTTP server with configurable response delays.
// Useful for testing timeout and rate limiting behavior without real delays.
type MockServer struct {
	server        *httptest.Server
	mu            sync.RWMutex
	handlers      []http.Handler
	responseDelay time.Duration
}

// NewMockServer creates a MockServer with no delay.
func NewMockServer() *MockServer {
	return &MockServer{
		responseDelay: 0,
	}
}

// WithHandler adds a handler to the mock server.
func (m *MockServer) WithHandler(handler http.Handler) *MockServer {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
	return m
}

// WithHandlerFunc adds a handler function to the mock server.
func (m *MockServer) WithHandlerFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *MockServer {
	return m.WithHandler(http.HandlerFunc(handler))
}

// WithResponseDelay sets a fixed delay for all responses.
func (m *MockServer) WithResponseDelay(delay time.Duration) *MockServer {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseDelay = delay
	return m
}

// URL returns the server's URL.
func (m *MockServer) URL() string {
	return m.server.URL
}

// Start begins serving requests.
func (m *MockServer) Start() {
	mux := http.NewServeMux()
	m.mu.RLock()
	handlers := m.handlers
	delay := m.responseDelay
	m.mu.RUnlock()

	for _, h := range handlers {
		mux.Handle("/", h)
	}

	mux.HandleFunc("/delay", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	})

	m.server = httptest.NewServer(mux)
}

// Stop shuts down the server.
func (m *MockServer) Stop() {
	if m.server != nil {
		m.server.Close()
	}
}

// LLMMock provides a mock server that simulates OpenAI/OpenRouter API.
type LLMMock struct {
	*MockServer
	responseContent string
	responseCode    int
	callCount       int
	mu              sync.RWMutex
}

// NewLLMMock creates a mock LLM server.
func NewLLMMock() *LLMMock {
	mock := &LLMMock{
		MockServer:      NewMockServer(),
		responseContent: `{"choices":[{"message":{"content":"Mock response"}}]}`,
		responseCode:    200,
	}
	mock.WithHandlerFunc("/v1/chat/completions", mock.handleChat)
	mock.WithHandlerFunc("/embeddings", mock.handleEmbeddings)
	return mock
}

// WithResponse sets the response content and status code.
func (l *LLMMock) WithResponse(content string, code int) *LLMMock {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.responseContent = content
	l.responseCode = code
	return l
}

// CallCount returns the number of requests received.
func (l *LLMMock) CallCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.callCount
}

func (l *LLMMock) handleChat(w http.ResponseWriter, r *http.Request) {
	l.mu.Lock()
	l.callCount++
	content := l.responseContent
	code := l.responseCode
	l.mu.Unlock()

	// Add configurable delay
	l.mu.RLock()
	delay := l.responseDelay
	l.mu.RUnlock()
	if delay > 0 {
		time.Sleep(delay)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = fmt.Fprint(w, content)
}

func (l *LLMMock) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	l.mu.Lock()
	l.callCount++
	l.mu.Unlock()

	l.mu.RLock()
	delay := l.responseDelay
	l.mu.RUnlock()
	if delay > 0 {
		time.Sleep(delay)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"data":[{"embedding":[0.1,0.2,0.3]}]}`)
}

// QdrantMock provides a mock server that simulates Qdrant API.
type QdrantMock struct {
	*MockServer
	collections map[string]bool
	points      map[string][]float32
	callCount   int
	mu          sync.RWMutex
}

// NewQdrantMock creates a mock Qdrant server.
func NewQdrantMock() *QdrantMock {
	mock := &QdrantMock{
		MockServer:  NewMockServer(),
		collections: make(map[string]bool),
		points:      make(map[string][]float32),
	}
	mock.WithHandlerFunc("/collections", mock.handleCollections)
	mock.WithHandlerFunc("/points", mock.handlePoints)
	mock.WithHandlerFunc("/search", mock.handleSearch)
	return mock
}

func (q *QdrantMock) handleCollections(w http.ResponseWriter, r *http.Request) {
	q.mu.Lock()
	q.callCount++
	q.mu.Unlock()

	q.mu.RLock()
	delay := q.responseDelay
	q.mu.RUnlock()
	if delay > 0 {
		time.Sleep(delay)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"result":{"collections":[]}}`)
}

func (q *QdrantMock) handlePoints(w http.ResponseWriter, r *http.Request) {
	q.mu.Lock()
	q.callCount++
	q.mu.Unlock()

	q.mu.RLock()
	delay := q.responseDelay
	q.mu.RUnlock()
	if delay > 0 {
		time.Sleep(delay)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"result":{"status":"ok"}}`)
}

func (q *QdrantMock) handleSearch(w http.ResponseWriter, r *http.Request) {
	q.mu.Lock()
	q.callCount++
	q.mu.Unlock()

	q.mu.RLock()
	delay := q.responseDelay
	q.mu.RUnlock()
	if delay > 0 {
		time.Sleep(delay)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"result":[{"id":"1","score":0.9}]}`)
}

// CallCount returns the number of requests received.
func (q *QdrantMock) CallCount() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.callCount
}
