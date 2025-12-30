package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestPublicEndpoints_BasicFlow(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "pub@example.com")
	create := map[string]any{"name": "Pub Bot", "language": "tr-TR"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID, nil)
	resG, _ := http.DefaultClient.Do(reqG)
	if resG.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resG.StatusCode)
	}
	resG.Body.Close()

	cr := chatReq{Message: "selam", SessionID: "s_pub"}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/public/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := http.DefaultClient.Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resCh.StatusCode)
	}
	resCh.Body.Close()
}
