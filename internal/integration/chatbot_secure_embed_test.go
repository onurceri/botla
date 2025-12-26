package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/pkg/policy"
)

func TestChatbot_SecureEmbed_UpdateAndGet(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "secure@example.com")

	// Enable secure embed for free plan to allow testing
	_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code=$1`, policy.PlanFree.String())
	if err != nil {
		t.Fatalf("failed to update plan config: %v", err)
	}

	create := map[string]any{"name": "Secure Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var created chatbot
	json.NewDecoder(resC.Body).Decode(&created)
	resC.Body.Close()

	// find via list
	reqL, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots", nil)
	reqL.Header.Set("Authorization", "Bearer "+token)
	resL, _ := http.DefaultClient.Do(reqL)
	if resL.StatusCode != http.StatusOK {
		t.Fatalf("list expected 200, got %d", resL.StatusCode)
	}
	var items []map[string]any
	json.NewDecoder(resL.Body).Decode(&items)
	resL.Body.Close()
	if len(items) == 0 {
		t.Fatalf("no chatbots listed")
	}
	id := items[0]["id"].(string)

	upd := map[string]any{"secure_embed_enabled": true, "allowed_domains": "example.com,another.com", "embed_secret": "SECRET123"}
	ub, _ := json.Marshal(upd)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+id, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resU.StatusCode)
	}
	var bot2 map[string]any
	json.NewDecoder(resU.Body).Decode(&bot2)
	resU.Body.Close()
	if v, ok := bot2["secure_embed_enabled"].(bool); !ok || !v {
		t.Fatalf("secure_embed_enabled not true")
	}

	reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+id, nil)
	reqG.Header.Set("Authorization", "Bearer "+token)
	resG, _ := http.DefaultClient.Do(reqG)
	if resG.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resG.StatusCode)
	}
	var bot3 map[string]any
	json.NewDecoder(resG.Body).Decode(&bot3)
	resG.Body.Close()
	if v, ok := bot3["secure_embed_enabled"].(bool); !ok || !v {
		t.Fatalf("secure_embed_enabled not persisted")
	}
}
