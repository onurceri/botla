package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"
)

func TestChat_Fallback_NoContext(t *testing.T) {
    oai := startOpenAIStub()
    t.Setenv("OPENAI_API_BASE", oai.URL)
    t.Setenv("QDRANT_URL", "")
    te, err := SetupTestEnv()
    if err != nil { t.Fatalf("setup failed: %v", err) }
    defer TeardownTestEnv(te)
    defer oai.Close()
    token := authToken(t, te.Server.URL, "nofallback@example.com")

    create := map[string]any{"name": "FB Bot"}
    cbj, _ := json.Marshal(create)
    reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
    reqC.Header.Set("Authorization", "Bearer "+token)
    reqC.Header.Set("Content-Type", "application/json")
    resC, _ := http.DefaultClient.Do(reqC)
    var bot chatbot
    json.NewDecoder(resC.Body).Decode(&bot)
    resC.Body.Close()

    cr := chatReq{Message: "merhaba", SessionID: "s2"}
    crb, _ := json.Marshal(cr)
    reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
    reqCh.Header.Set("Authorization", "Bearer "+token)
    reqCh.Header.Set("Content-Type", "application/json")
    resCh, _ := http.DefaultClient.Do(reqCh)
    if resCh.StatusCode != http.StatusOK { t.Fatalf("expected 200, got %d", resCh.StatusCode) }
    var crp chatResp
    json.NewDecoder(resCh.Body).Decode(&crp)
    resCh.Body.Close()
    if crp.Response != "Yeterli bilgi bulamadım." || crp.TokensUsed != 0 { t.Fatalf("expected fallback, got %q/%d", crp.Response, crp.TokensUsed) }
}

