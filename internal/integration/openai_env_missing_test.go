package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"
)

func TestChat_OpenAIEnvMissing_500(t *testing.T) {
    oai := startOpenAIStub()
    qd := startQdrantStub()
    t.Setenv("OPENAI_API_BASE", oai.URL)
    t.Setenv("QDRANT_URL", qd.URL)
    te, err := SetupTestEnv()
    if err != nil { t.Fatalf("setup failed: %v", err) }
    defer TeardownTestEnv(te)
    defer oai.Close()
    defer qd.Close()
    // Ensure OPENAI_API_KEY is missing to trigger server-side 500
    t.Setenv("OPENAI_API_KEY", "")

    token := authToken(t, te.Server.URL, "envmissing@example.com")
    create := map[string]any{"name": "Env Bot"}
    cbj, _ := json.Marshal(create)
    reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
    reqC.Header.Set("Authorization", "Bearer "+token)
    reqC.Header.Set("Content-Type", "application/json")
    resC, _ := http.DefaultClient.Do(reqC)
    var bot chatbot
    json.NewDecoder(resC.Body).Decode(&bot)
    resC.Body.Close()

    cr := chatReq{Message: "merhaba", SessionID: "s-env"}
    crb, _ := json.Marshal(cr)
    reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
    reqCh.Header.Set("Authorization", "Bearer "+token)
    reqCh.Header.Set("Content-Type", "application/json")
    resCh, _ := http.DefaultClient.Do(reqCh)
    if resCh.StatusCode != http.StatusInternalServerError { t.Fatalf("expected 500, got %d", resCh.StatusCode) }
    resCh.Body.Close()
}

func TestChat_QdrantEnvMissing_Fallback(t *testing.T) {
    // Missing QDRANT_URL should result in contextless chat fallback
    oai := startOpenAIStub()
    t.Setenv("OPENAI_API_BASE", oai.URL)
    t.Setenv("QDRANT_URL", "")
    te, err := SetupTestEnv()
    if err != nil { t.Fatalf("setup failed: %v", err) }
    defer TeardownTestEnv(te)
    defer oai.Close()

    token := authToken(t, te.Server.URL, "qdenv@example.com")
    create := map[string]any{"name": "QD Env Bot"}
    cbj, _ := json.Marshal(create)
    reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
    reqC.Header.Set("Authorization", "Bearer "+token)
    reqC.Header.Set("Content-Type", "application/json")
    resC, _ := http.DefaultClient.Do(reqC)
    var bot chatbot
    json.NewDecoder(resC.Body).Decode(&bot)
    resC.Body.Close()

    cr := chatReq{Message: "merhaba", SessionID: "s-qdenv"}
    crb, _ := json.Marshal(cr)
    reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
    reqCh.Header.Set("Authorization", "Bearer "+token)
    reqCh.Header.Set("Content-Type", "application/json")
    resCh, _ := http.DefaultClient.Do(reqCh)
    if resCh.StatusCode != http.StatusOK { t.Fatalf("expected 200, got %d", resCh.StatusCode) }
    var crp chatResp
    json.NewDecoder(resCh.Body).Decode(&crp)
    resCh.Body.Close()
    if crp.Response != "Yeterli bilgi bulamadım." || crp.TokensUsed != 0 { t.Fatalf("expected qdrant-env fallback, got %q/%d", crp.Response, crp.TokensUsed) }
}
