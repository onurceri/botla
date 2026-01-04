package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/lib/pq"
	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/policy"
)

type CapturedLLMRequest struct {
	Model       string  `json:"model"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

type LLMStub struct {
	Server   *httptest.Server
	Requests []CapturedLLMRequest
	Mu       sync.Mutex
}

func startLLMStub() *LLMStub {
	s := &LLMStub{}
	h := http.NewServeMux()

	// OpenAI/Compatible Chat Completions Handler
	h.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req CapturedLLMRequest
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

	// Embeddings Handler
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

	s.Server = httptest.NewServer(h)
	return s
}

func TestChat_TemperatureConfiguration(t *testing.T) {
	stub := startLLMStub()
	defer stub.Server.Close()

	qd := fixtures.StartQdrantStub()
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

	token := authToken(t, te.Server.URL, "temp_test@example.com")

	models := []string{"gpt-4o", "gpt-4o-mini", "anthropic:claude-3-5-sonnet-20241022", "google:gemini-1.5-flash", "openrouter:meta-llama/llama-3"}
	_, _ = te.DB.Exec(`UPDATE plan_limits SET chat_allowed_models = $1, chat_max_monthly_tokens = 0 WHERE plan_id = (SELECT id FROM plans WHERE code = $2)`,
		pq.Array(models), policy.PlanFree.String())
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), "temp_test@example.com")

	testCases := []struct {
		name         string
		temperature  *float32
		maxTokens    *int
		expectedTemp float32
		expectedMT   int
	}{
		{"Chatbot Temp 1.0, MaxTokens 2048", float32Ptr(1.0), intPtr(2048), 1.0, 2048},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Chatbot
			create := map[string]any{
				"name": "Temp Bot " + tc.name,
			}
			if tc.temperature != nil {
				create["temperature"] = *tc.temperature
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
				ID string `json:"id"`
			}
			json.NewDecoder(resC.Body).Decode(&bot)
			resC.Body.Close()

			// Send Chat - use a unique session to avoid action loop issues
			cr := map[string]string{"message": "hello", "session_id": "s-temp-" + tc.name}
			crb, _ := json.Marshal(cr)
			reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
			reqCh.Header.Set("Authorization", "Bearer "+token)
			reqCh.Header.Set("Content-Type", "application/json")
			resCh, _ := testHTTPClient().Do(reqCh)
			if resCh.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resCh.Body)
				t.Fatalf("chat failed: %d - %s", resCh.StatusCode, string(body))
			}
			resCh.Body.Close()

			// Verify Stub
			stub.Mu.Lock()
			if len(stub.Requests) == 0 {
				stub.Mu.Unlock()
				t.Fatal("no requests received by LLM stub")
			}
			lastReq := stub.Requests[len(stub.Requests)-1]
			stub.Mu.Unlock()

			t.Logf("Captured LLM Request - Temp: %f, MaxTokens: %d, Model: %s", lastReq.Temperature, lastReq.MaxTokens, lastReq.Model)

			// Verify temperature
			if diff := lastReq.Temperature - tc.expectedTemp; diff < -0.001 || diff > 0.001 {
				t.Errorf("expected temperature %f, got %f", tc.expectedTemp, lastReq.Temperature)
			}

			// Verify max_tokens (only if expected > 0)
			if tc.expectedMT > 0 && lastReq.MaxTokens != tc.expectedMT {
				t.Errorf("expected max_tokens %d, got %d", tc.expectedMT, lastReq.MaxTokens)
			}
		})
	}
}

func TestChat_MaxTokensConfiguration(t *testing.T) {
	stub := startLLMStub()
	defer stub.Server.Close()

	qd := fixtures.StartQdrantStub()
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

	token := authToken(t, te.Server.URL, "maxtokens_test@example.com")

	models := []string{"gpt-4o", "gpt-4o-mini", "anthropic:claude-3-5-sonnet-20241022"}
	_, _ = te.DB.Exec(`UPDATE plan_limits SET chat_allowed_models = $1, chat_max_monthly_tokens = 0 WHERE plan_id = (SELECT id FROM plans WHERE code = $2)`,
		pq.Array(models), policy.PlanFree.String())
	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), "maxtokens_test@example.com")

	testCases := []struct {
		name       string
		maxTokens  *int
		expectedMT int
	}{
		{"MaxTokens 256", intPtr(256), 256},
		{"MaxTokens 1024", intPtr(1024), 1024},
		{"MaxTokens 4096", intPtr(4096), 4096},
		{"Default MaxTokens", nil, 4096},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Chatbot with default temperature 0.5
			create := map[string]any{
				"name":        "MaxTokens Bot " + tc.name,
				"temperature": 0.5,
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
				ID string `json:"id"`
			}
			json.NewDecoder(resC.Body).Decode(&bot)
			resC.Body.Close()

			// Send Chat
			cr := map[string]string{"message": "hello", "session_id": "s-maxtokens-" + tc.name}
			crb, _ := json.Marshal(cr)
			reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
			reqCh.Header.Set("Authorization", "Bearer "+token)
			reqCh.Header.Set("Content-Type", "application/json")
			resCh, _ := testHTTPClient().Do(reqCh)
			if resCh.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resCh.Body)
				t.Fatalf("chat failed: %d - %s", resCh.StatusCode, string(body))
			}
			resCh.Body.Close()

			// Verify Stub
			stub.Mu.Lock()
			if len(stub.Requests) == 0 {
				stub.Mu.Unlock()
				t.Fatal("no requests received by LLM stub")
			}
			lastReq := stub.Requests[len(stub.Requests)-1]
			stub.Mu.Unlock()

			t.Logf("Captured LLM Request - Temp: %f, MaxTokens: %d", lastReq.Temperature, lastReq.MaxTokens)

			// Verify temperature is 0.5 (default)
			if diff := lastReq.Temperature - 0.5; diff < -0.001 || diff > 0.001 {
				t.Errorf("expected temperature 0.5, got %f", lastReq.Temperature)
			}

			// Verify max_tokens
			if lastReq.MaxTokens != tc.expectedMT {
				t.Errorf("expected max_tokens %d, got %d", tc.expectedMT, lastReq.MaxTokens)
			}
		})
	}
}
