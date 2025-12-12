package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func startQdrantErrorStub() *httptest.Server {
	h := http.NewServeMux()
	h.HandleFunc("/collections/embeddings", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h.HandleFunc("/collections/embeddings/points/search", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	return httptest.NewServer(h)
}

func TestChat_QdrantSearchError_Fallback(t *testing.T) {
	oai := startOpenAIStub()
	qd := startQdrantErrorStub()
	t.Setenv("OPENAI_API_BASE", oai.URL)
	t.Setenv("QDRANT_URL", qd.URL)
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	token := authToken(t, te.Server.URL, "qderr@example.com")
	// Create bot with static fallback mode to test 0-token fallback
	create := map[string]any{
		"name": "QD Err Bot",
		"threshold_config": map[string]any{
			"high_threshold":          0.50,
			"medium_threshold":        0.30,
			"fallback_mode":           "static",
			"show_confidence_warning": false,
		},
	}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	cr := chatReq{Message: "selam", SessionID: "s4"}
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
	if crp.Response != "Yeterli bilgi bulamadım." || crp.TokensUsed != 0 {
		t.Fatalf("expected fallback, got %q/%d", crp.Response, crp.TokensUsed)
	}
}
