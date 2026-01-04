package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/pkg/config"
)

func TestChat_DefaultTurkishPrompt(t *testing.T) {
	t.Parallel()
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.QDRANT_URL = qd.URL
		cfg.RAG_TOPK = 3
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	token := authToken(t, te.Server.URL, "prompttr@example.com")

	// Get workspace ID for the user
	var wsID string
	err = te.DB.QueryRow(`SELECT w.id FROM workspaces w 
		JOIN organizations o ON w.organization_id = o.id 
		JOIN memberships m ON o.id = m.organization_id 
		JOIN users u ON m.user_id = u.id 
		WHERE u.email = 'prompttr@example.com' LIMIT 1`).Scan(&wsID)
	if err != nil {
		t.Fatalf("failed to get workspace id: %v", err)
	}

	create := map[string]any{"name": "TR Prompt Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	reqC.Header.Set("X-Workspace-ID", wsID)
	resC, _ := testHTTPClient().Do(reqC)
	if resC.StatusCode != http.StatusCreated {
		t.Fatalf("create bot failed: %d", resC.StatusCode)
	}
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// ensure system_prompt is empty so default applies
	cr := chatReq{Message: "merhaba", SessionID: "s-tr"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	reqCh.Header.Set("X-Workspace-ID", wsID)
	resCh, _ := testHTTPClient().Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resCh.StatusCode)
	}
	var crp chatResp
	json.NewDecoder(resCh.Body).Decode(&crp)
	resCh.Body.Close()
	if crp.Response == "" {
		t.Fatalf("expected non-empty response with default Turkish prompt")
	}
}
