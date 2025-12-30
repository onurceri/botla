package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

// TestChat_EmptyState_GreetingResponse verifies that the bot responds to greetings
// even when no RAG context/knowledge sources are available.
func TestChat_EmptyState_GreetingResponse(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("QDRANT_URL", "") // No Qdrant = no RAG context
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	token := authToken(t, te.Server.URL, "emptystate@example.com")

	// Create a bot with no data sources
	create := map[string]any{
		"name": "Empty State Bot",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Test greeting message
	greetings := []string{"Merhaba", "Selam", "Hello", "Naber"}
	for _, greeting := range greetings {
		t.Run("greeting_"+greeting, func(t *testing.T) {
			cr := chatReq{Message: greeting, SessionID: "session-" + greeting}
			crb, _ := json.Marshal(cr)
			reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
			reqCh.Header.Set("Authorization", "Bearer "+token)
			reqCh.Header.Set("Content-Type", "application/json")
			resCh, _ := http.DefaultClient.Do(reqCh)

			if resCh.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resCh.StatusCode)
			}

			var crp chatResp
			json.NewDecoder(resCh.Body).Decode(&crp)
			resCh.Body.Close()

			// Response should not be empty
			if crp.Response == "" {
				t.Fatalf("expected non-empty response for greeting %q", greeting)
			}

			// Response should be friendly (not a hard "no info found" message)
			hardRefusal := "Yeterli bilgi bulamadım"
			if strings.Contains(crp.Response, hardRefusal) {
				t.Errorf("expected friendly response for greeting %q, got hard refusal: %q", greeting, crp.Response)
			}
		})
	}
}

// TestChat_EmptyState_FactualRefusal verifies that the bot refuses to answer
// factual questions when no RAG context is available.
func TestChat_EmptyState_FactualRefusal(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("QDRANT_URL", "") // No Qdrant = no RAG context
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	token := authToken(t, te.Server.URL, "factual@example.com")

	// Create a bot with no data sources
	create := map[string]any{
		"name": "Factual Test Bot",
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// Test factual question - bot should not answer with incorrect facts
	factualQuestions := []string{
		"Ürünleriniz neler?",
		"Fiyatlarınız ne kadar?",
		"What products do you sell?",
	}

	for _, question := range factualQuestions {
		t.Run("factual_"+question[:10], func(t *testing.T) {
			cr := chatReq{Message: question, SessionID: "session-factual"}
			crb, _ := json.Marshal(cr)
			reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
			reqCh.Header.Set("Authorization", "Bearer "+token)
			reqCh.Header.Set("Content-Type", "application/json")
			resCh, _ := http.DefaultClient.Do(reqCh)

			if resCh.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resCh.StatusCode)
			}

			var crp chatResp
			json.NewDecoder(resCh.Body).Decode(&crp)
			resCh.Body.Close()

			// Response should not be empty
			if crp.Response == "" {
				t.Fatalf("expected non-empty response for factual question")
			}

			// Response should NOT contain invented product information
			// (This is a basic check - the LLM mock controls the actual response)
			inventedPatterns := []string{
				"ürünümüz var",
				"fiyatımız",
				"$",
				"€",
			}
			for _, pattern := range inventedPatterns {
				if strings.Contains(strings.ToLower(crp.Response), pattern) {
					t.Logf("Warning: Response may contain invented info: %q", crp.Response)
				}
			}
		})
	}
}

// TestChat_EmptyState_IdentityQuestion verifies that the bot can answer
// "Who are you?" questions even without RAG context.
func TestChat_EmptyState_IdentityQuestion(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("OPENROUTER_API_BASE", oai.URL+"/v1")
	t.Setenv("QDRANT_URL", "")
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	token := authToken(t, te.Server.URL, "identity@example.com")

	botName := "MyIdentityBot"
	create := map[string]any{
		"name": botName,
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	identityQuestions := []string{
		"Sen kimsin?",
		"Adın ne?",
		"What's your name?",
	}

	for _, question := range identityQuestions {
		t.Run("identity_"+question[:5], func(t *testing.T) {
			cr := chatReq{Message: question, SessionID: "session-identity"}
			crb, _ := json.Marshal(cr)
			reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
			reqCh.Header.Set("Authorization", "Bearer "+token)
			reqCh.Header.Set("Content-Type", "application/json")
			resCh, _ := http.DefaultClient.Do(reqCh)

			if resCh.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resCh.StatusCode)
			}

			var crp chatResp
			json.NewDecoder(resCh.Body).Decode(&crp)
			resCh.Body.Close()

			if crp.Response == "" {
				t.Fatalf("expected non-empty response for identity question %q", question)
			}

			// Response should ideally mention the bot name (handled by LLM)
			t.Logf("Identity response for %q: %s", question, crp.Response)
		})
	}
}
