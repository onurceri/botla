package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/pkg/config"
)

func TestAnalytics_UpdatesAfterChat(t *testing.T) {
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

	token, err := te.AuthToken("anupd@example.com")
	if err != nil {
		t.Fatalf("auth token failed: %v", err)
	}

	// create chatbot
	create := map[string]any{"name": "AN Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// perform chat (user + assistant messages)
	cr := chatReq{Message: "merhaba", SessionID: "s5"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := testHTTPClient().Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resCh.StatusCode)
	}
	resCh.Body.Close()

	// read analytics and assert totals increased with retry
	var totalM, totalC int
	found := false
	for i := 0; i < 10; i++ {
		reqA, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/analytics", nil)
		reqA.Header.Set("Authorization", "Bearer "+token)
		resA, _ := testHTTPClient().Do(reqA)
		if resA.StatusCode == http.StatusOK {
			var series []struct {
				Date          string
				Messages      int
				Conversations int
			}
			json.NewDecoder(resA.Body).Decode(&series)
			resA.Body.Close()
			totalM, totalC = 0, 0
			for _, p := range series {
				totalM += p.Messages
				totalC += p.Conversations
			}
		}

		if totalM >= 2 && totalC >= 1 {
			found = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if !found {
		t.Fatalf("expected totals >= (2 messages, 1 conv), got (%d, %d)", totalM, totalC)
	}
}
