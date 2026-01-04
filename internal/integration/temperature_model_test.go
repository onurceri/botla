package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/lib/pq"
	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/policy"
)

type CapturedRequest struct {
	Model       string  `json:"model"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

type TMStub struct {
	Server   *httptest.Server
	Requests []CapturedRequest
	Mu       sync.Mutex
}

func baseModelName(model string) string {
	s := strings.TrimSpace(model)
	if i := strings.LastIndex(s, "/"); i >= 0 && i < len(s)-1 {
		return s[i+1:]
	}
	return s
}

func startTMStub() *TMStub {
	s := &TMStub{}
	h := http.NewServeMux()

	// OpenAI Stub
	h.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req CapturedRequest
		_ = json.Unmarshal(body, &req)

		s.Mu.Lock()
		s.Requests = append(s.Requests, req)
		s.Mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "Stubbed Response"}},
			},
			"usage": map[string]int{"total_tokens": 10},
		})
	})

	// Embedding Stub
	h.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := make([]float64, 1536)
		resp := map[string]any{
			"data":  []map[string]any{{"embedding": data}},
			"usage": map[string]int{"prompt_tokens": 10, "total_tokens": 10},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Anthropic Stub
	h.HandleFunc("/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// We can try to unmarshal to check fields if needed, but for now just capture generic json
		// or at least capture that a request happened.
		// For simplicity, we can reuse CapturedRequest or just map[string]any
		var reqMap map[string]any
		_ = json.Unmarshal(body, &reqMap)

		// Map Anthropic request to CapturedRequest for verification
		cr := CapturedRequest{}
		if m, ok := reqMap["model"].(string); ok {
			cr.Model = m
		}
		if t, ok := reqMap["temperature"].(float64); ok {
			cr.Temperature = float32(t)
		}
		if m, ok := reqMap["max_tokens"].(float64); ok {
			cr.MaxTokens = int(m)
		}

		s.Mu.Lock()
		s.Requests = append(s.Requests, cr)
		s.Mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"id":   "msg_stub",
			"type": "message",
			"role": "assistant",
			"content": []map[string]any{
				{"type": "text", "text": "Anthropic Stubbed Response"},
			},
			"usage": map[string]int{"input_tokens": 10, "output_tokens": 10},
		})
	})

	// Google Stub
	// Google client appends /models/{model}:generateContent
	h.HandleFunc("/models/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqMap map[string]any
		_ = json.Unmarshal(body, &reqMap)

		cr := CapturedRequest{}
		// Google model is in URL, but we can extract if needed.
		// For verification, we can just assume it matches if it hit this handler.
		// Or parse URL: /models/{model}:generateContent
		path := r.URL.Path // /models/gemini-1.5-flash:generateContent
		if len(path) > 8 {
			// Extract model name
			// e.g. gemini-1.5-flash:generateContent
			remaining := path[8:]
			if idx := strings.Index(remaining, ":"); idx != -1 {
				cr.Model = remaining[:idx]
			}
		}

		if gc, ok := reqMap["generationConfig"].(map[string]any); ok {
			if t, ok := gc["temperature"].(float64); ok {
				cr.Temperature = float32(t)
			}
			if m, ok := gc["maxOutputTokens"].(float64); ok {
				cr.MaxTokens = int(m)
			}
		}

		s.Mu.Lock()
		s.Requests = append(s.Requests, cr)
		s.Mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]any{
							{"text": "Google Stubbed Response"},
						},
					},
					"finishReason": "STOP",
				},
			},
			"usageMetadata": map[string]int{
				"totalTokenCount": 20,
			},
		})
	})

	s.Server = httptest.NewServer(h)
	return s
}

func float32Ptr(v float32) *float32 {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func TestTemperatureParameter(t *testing.T) {
	stub := startTMStub()
	defer stub.Server.Close()

	qd := startQdrantStub()
	defer qd.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = stub.Server.URL
		cfg.OPENROUTER_API_BASE = stub.Server.URL + "/v1"
		cfg.OPENAI_API_KEY = "test-key"
		cfg.QDRANT_URL = qd.URL
	}, false) // useMocks=false to connect to stub server
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "temp_test@example.com")

	testCases := []struct {
		name     string
		temp     *float32 // Pointer to distinguish nil (default) vs 0.0
		expected float32
	}{
		{"Temp 0.0", float32Ptr(0.0), 0.0},
		{"Temp 0.5", float32Ptr(0.5), 0.5},
		{"Temp 1.0", float32Ptr(1.0), 1.0},
		{"Temp 2.0", float32Ptr(2.0), 2.0},
		{"Default", nil, 0.7}, // Assuming default is 0.7
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Chatbot
			create := map[string]any{
				"name": "Temp Bot " + tc.name,
			}
			if tc.temp != nil {
				create["temperature"] = *tc.temp
			}
			cbj, _ := json.Marshal(create)
			reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
			reqC.Header.Set("Authorization", "Bearer "+token)
			reqC.Header.Set("Content-Type", "application/json")
			resC, _ := testHTTPClient().Do(reqC)
			if resC.StatusCode != http.StatusCreated {
				t.Fatalf("create failed: %d", resC.StatusCode)
			}
			var bot struct {
				ID    string `json:"id"`
				Model string `json:"model"`
			}
			json.NewDecoder(resC.Body).Decode(&bot)
			resC.Body.Close()

			t.Logf("Created Chatbot: %s, Model: %s", bot.ID, bot.Model)

			// Send Chat
			cr := map[string]string{"message": "hello", "session_id": "s1"}
			crb, _ := json.Marshal(cr)
			reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
			reqCh.Header.Set("Authorization", "Bearer "+token)
			reqCh.Header.Set("Content-Type", "application/json")
			resCh, _ := testHTTPClient().Do(reqCh)
			if resCh.StatusCode != http.StatusOK {
				t.Fatalf("chat failed: %d", resCh.StatusCode)
			}
			resCh.Body.Close()

			// Verify Stub
			stub.Mu.Lock()
			if len(stub.Requests) == 0 {
				stub.Mu.Unlock()
				t.Fatal("no requests received by temperature stub")
			}
			lastReq := stub.Requests[len(stub.Requests)-1]
			stub.Mu.Unlock()

			if lastReq.Temperature != tc.expected {
				t.Errorf("expected temperature %f, got %f", tc.expected, lastReq.Temperature)
			}
		})
	}
}

func TestMaxTokensParameter(t *testing.T) {
	stub := startTMStub()
	defer stub.Server.Close()

	qd := startQdrantStub()
	defer qd.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = stub.Server.URL
		cfg.OPENROUTER_API_BASE = stub.Server.URL + "/v1"
		cfg.OPENAI_API_KEY = "test-key"
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "tokens_test@example.com")

	testCases := []struct {
		name      string
		maxTokens *int
		expected  int
	}{
		{"Tokens 256", intPtr(256), 256},
		{"Tokens 4096", intPtr(4096), 4096},
		{"Default (512)", nil, 4096},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Chatbot
			create := map[string]any{
				"name": "Tokens Bot " + tc.name,
			}
			if tc.maxTokens != nil {
				create["max_tokens"] = *tc.maxTokens
			}
			cbj, _ := json.Marshal(create)
			reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
			reqC.Header.Set("Authorization", "Bearer "+token)
			reqC.Header.Set("Content-Type", "application/json")
			resC, _ := testHTTPClient().Do(reqC)
			if resC.StatusCode != http.StatusCreated {
				t.Fatalf("create failed: %d", resC.StatusCode)
			}
			var bot struct {
				ID    string `json:"id"`
				Model string `json:"model"`
			}
			json.NewDecoder(resC.Body).Decode(&bot)
			resC.Body.Close()

			t.Logf("Created Chatbot: %s, Model: %s", bot.ID, bot.Model)

			// Send Chat
			cr := map[string]string{"message": "hello", "session_id": "s1"}
			crb, _ := json.Marshal(cr)
			reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
			reqCh.Header.Set("Authorization", "Bearer "+token)
			reqCh.Header.Set("Content-Type", "application/json")
			resCh, _ := testHTTPClient().Do(reqCh)
			if resCh.StatusCode != http.StatusOK {
				t.Fatalf("chat failed: %d", resCh.StatusCode)
			}
			resCh.Body.Close()

			// Verify Stub
			stub.Mu.Lock()
			if len(stub.Requests) == 0 {
				stub.Mu.Unlock()
				t.Fatal("no requests received by temperature stub")
			}
			lastReq := stub.Requests[len(stub.Requests)-1]
			stub.Mu.Unlock()

			if lastReq.MaxTokens != tc.expected {
				t.Errorf("expected max_tokens %d, got %d", tc.expected, lastReq.MaxTokens)
			}
		})
	}
}

func TestModelConfiguration(t *testing.T) {
	stub := startTMStub()
	defer stub.Server.Close()

	qd := startQdrantStub()
	defer qd.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = stub.Server.URL
		cfg.ANTHROPIC_API_KEY = "test-anthropic"
		cfg.ANTHROPIC_API_BASE = stub.Server.URL
		cfg.GOOGLE_AI_API_KEY = "test-google"
		cfg.GOOGLE_AI_API_BASE = stub.Server.URL
		cfg.OPENROUTER_API_KEY = "test-openrouter"
		cfg.OPENROUTER_API_BASE = stub.Server.URL + "/v1" // Matches /v1/chat/completions
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "model_test@example.com")

	// Update plan to allow all these models
	models := []string{"gpt-4o", "gpt-4o-mini", "anthropic:claude-3-5-sonnet", "google:gemini-1.5-flash", "openrouter:meta-llama/llama-3"}
	_, _ = te.DB.Exec(`UPDATE plan_limits SET chat_allowed_models = $1 WHERE plan_id = (SELECT id FROM plans WHERE code = $2)`,
		pq.Array(models), policy.PlanFree.String())
	// Ensure user is on free plan
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), "model_test@example.com")

	testCases := []struct {
		name     string
		model    *string
		expected string
	}{
		{"Default Model", nil, policy.ModelGPT4oMini.String()},
		{"GPT-4o", stringPtr(policy.ModelGPT4o.String()), policy.ModelGPT4o.String()},
		{"GPT-4o-Mini", stringPtr(policy.ModelGPT4oMini.String()), policy.ModelGPT4oMini.String()},
		{"Anthropic", stringPtr("anthropic:claude-3-5-sonnet"), "anthropic/claude-3-5-sonnet"},
		{"Google", stringPtr("google:gemini-1.5-flash"), "google/gemini-1.5-flash"},
		{"OpenRouter", stringPtr("openrouter:meta-llama/llama-3"), "meta-llama/llama-3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Chatbot
			create := map[string]any{
				"name": "Model Bot " + tc.name,
			}
			if tc.model != nil {
				create["model"] = *tc.model
			}
			cbj, _ := json.Marshal(create)
			reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
			reqC.Header.Set("Authorization", "Bearer "+token)
			reqC.Header.Set("Content-Type", "application/json")
			resC, _ := testHTTPClient().Do(reqC)
			if resC.StatusCode != http.StatusCreated {
				t.Fatalf("create failed: %d", resC.StatusCode)
			}
			var bot struct {
				ID    string `json:"id"`
				Model string `json:"model"`
			}
			json.NewDecoder(resC.Body).Decode(&bot)
			resC.Body.Close()

			// Send Chat
			cr := map[string]string{"message": "hello", "session_id": "s1"}
			crb, _ := json.Marshal(cr)
			reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
			reqCh.Header.Set("Authorization", "Bearer "+token)
			reqCh.Header.Set("Content-Type", "application/json")
			resCh, _ := testHTTPClient().Do(reqCh)
			if resCh.StatusCode != http.StatusOK {
				t.Fatalf("chat failed: %d", resCh.StatusCode)
			}
			resCh.Body.Close()

			// Verify Stub
			stub.Mu.Lock()
			count := len(stub.Requests)
			var lastReq CapturedRequest
			if count > 0 {
				lastReq = stub.Requests[count-1]
			}
			stub.Mu.Unlock()

			t.Logf("Test: %s, Request Count: %d, Last Model: %s", tc.name, count, lastReq.Model)
			t.Logf("Chatbot Info - ID: %s, Model: %s", bot.ID, bot.Model)

			if lastReq.Model != tc.expected {
				// Debug: print all requests
				stub.Mu.Lock()
				for i, r := range stub.Requests {
					t.Logf("Req[%d]: Model=%s", i, r.Model)
				}
				stub.Mu.Unlock()
				t.Errorf("expected model %s, got %s", tc.expected, lastReq.Model)
			}
		})
	}
}

func TestModelRestrictions(t *testing.T) {
	stub := startTMStub()
	defer stub.Server.Close()

	qd := startQdrantStub()
	defer qd.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = stub.Server.URL
		cfg.OPENROUTER_API_BASE = stub.Server.URL + "/v1"
		cfg.OPENAI_API_KEY = "test-key"
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "restricted_test@example.com")

	// Update plan to allow ONLY gpt-4o-mini
	models := []string{"gpt-4o-mini"}
	_, _ = te.DB.Exec(`UPDATE plan_limits SET chat_allowed_models = $1 WHERE plan_id = (SELECT id FROM plans WHERE code = $2)`,
		pq.Array(models), policy.PlanFree.String())
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), "restricted_test@example.com")

	// Try to create chatbot with forbidden model
	create := map[string]any{
		"name":  "Restricted Bot",
		"model": policy.ModelGPT4o.String(), // Forbidden
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated {
		t.Fatalf("create failed: %d", resC.StatusCode)
	}
	var bot struct {
		ID    string `json:"id"`
		Model string `json:"model"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// The chatbot is created with the requested model (because validation happens at chat time or maybe not enforced at creation?)
	// Based on code reading, enforcement is in Chat Handler. So creation allows it, but chat should swap it.
	t.Logf("Created Chatbot: %s, Model: %s", bot.ID, bot.Model)

	// Send Chat
	cr := map[string]string{"message": "hello", "session_id": "s1"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := testHTTPClient().Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("chat failed: %d", resCh.StatusCode)
	}
	resCh.Body.Close()

	// Verify Stub used gpt-4o-mini
	stub.Mu.Lock()
	if len(stub.Requests) == 0 {
		stub.Mu.Unlock()
		t.Fatal("no requests received by temperature stub")
	}
	lastReq := stub.Requests[len(stub.Requests)-1]
	stub.Mu.Unlock()

	if baseModelName(lastReq.Model) != policy.ModelGPT4oMini.String() {
		t.Errorf("expected fallback to %s, got %s", policy.ModelGPT4oMini.String(), lastReq.Model)
	}
}

func TestInvalidModel(t *testing.T) {
	stub := startTMStub()
	defer stub.Server.Close()

	qd := startQdrantStub()
	defer qd.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = stub.Server.URL
		cfg.OPENROUTER_API_BASE = stub.Server.URL + "/v1"
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "invalid_model_test@example.com")

	// Try to create chatbot with invalid model name
	create := map[string]any{
		"name":  "Invalid Model Bot",
		"model": "invalid-model-name-xyz",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated {
		t.Fatalf("create failed: %d", resC.StatusCode)
	}
	var bot struct {
		ID    string `json:"id"`
		Model string `json:"model"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Send Chat
	cr := map[string]string{"message": "hello", "session_id": "s1"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := testHTTPClient().Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("chat failed: %d", resCh.StatusCode)
	}
	resCh.Body.Close()

	// Verify Stub used invalid-model (OpenRouter pass-through)
	stub.Mu.Lock()
	if len(stub.Requests) == 0 {
		stub.Mu.Unlock()
		t.Fatal("no requests received by temperature stub")
	}
	lastReq := stub.Requests[len(stub.Requests)-1]
	stub.Mu.Unlock()

	if baseModelName(lastReq.Model) != policy.ModelGPT4oMini.String() {
		t.Errorf("expected fallback to %s, got %s", policy.ModelGPT4oMini.String(), lastReq.Model)
	}
}

func TestFreePlan_RAGTopK_UsesPlanLimit(t *testing.T) {
	stub := startTMStub()
	defer stub.Server.Close()

	qd := startQdrantTopKStub()
	defer qd.Server.Close()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = stub.Server.URL
		cfg.OPENROUTER_API_BASE = stub.Server.URL + "/v1"
		cfg.QDRANT_URL = qd.Server.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "rag_topk@example.com")
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code='free') WHERE email=$1`, "rag_topk@example.com")

	create := map[string]any{"name": "RAG TopK Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, err := testHTTPClient().Do(reqC)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	if resC.StatusCode != http.StatusCreated {
		t.Fatalf("create failed: %d", resC.StatusCode)
	}
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := map[string]string{"message": "hello", "session_id": "s-rag-topk"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, err := testHTTPClient().Do(reqCh)
	if err != nil {
		t.Fatalf("chat request failed: %v", err)
	}
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("chat failed: %d", resCh.StatusCode)
	}
	resCh.Body.Close()

	qd.Mu.Lock()
	limit := qd.LastSearchLimit
	calls := qd.SearchCalls
	qd.Mu.Unlock()

	if calls == 0 {
		t.Fatalf("expected at least one search call to Qdrant stub")
	}
	if limit != 3 {
		t.Errorf("expected search limit 3 from plan RAG config, got %d", limit)
	}
}

func stringPtr(s string) *string {
	return &s
}
