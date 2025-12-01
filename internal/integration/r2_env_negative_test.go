package integration

import (
    "bytes"
    "encoding/json"
    "io"
    "mime/multipart"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "github.com/onurceri/botla-co/internal/api/handlers"
    "github.com/onurceri/botla-co/pkg/logger"
    "github.com/onurceri/botla-co/pkg/middleware"
    "github.com/onurceri/botla-co/pkg/storage"
)

// Build a minimal mux that uses R2 storage to exercise upload paths without hitting external endpoints
func buildR2Mux(te *TestEnv, bucket string) http.Handler {
    mux := http.NewServeMux()
    ah := &handlers.AuthHandlers{DB: te.DB, Secret: te.Cfg.JWT_SECRET}
    mux.HandleFunc("/api/v1/auth/register", ah.RegisterHandler)
    mux.HandleFunc("/api/v1/auth/login", ah.LoginHandler)
    ch := &handlers.ChatbotHandlers{DB: te.DB}
    mux.Handle("/api/v1/chatbots", middleware.AuthMiddleware(te.Cfg.JWT_SECRET)(http.HandlerFunc(ch.ListOrCreate)))
    // R2 storage with empty bucket to force client-side error without network leakage
    r2, _ := storage.NewR2Storage("acc-test", "AKIA_TEST_LEAK", "SECRET_TEST_LEAK", bucket)
    sh := &handlers.SourcesHandlers{DB: te.DB, Storage: r2}
    chh := &handlers.ChatHandlers{DB: te.DB}
    mux.Handle("/api/v1/chatbots/", middleware.AuthMiddleware(te.Cfg.JWT_SECRET)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        const p = "/api/v1/chatbots/"
        if len(r.URL.Path) >= len(p) && r.URL.Path[len(r.URL.Path)-len("/sources"):] == "/sources" {
            sh.ChatbotSources(w, r)
            return
        }
        ch.ByID(w, r)
    })))
    mux.Handle("/api/v1/messages/", middleware.AuthMiddleware(te.Cfg.JWT_SECRET)(http.HandlerFunc(chh.FeedbackHandler)))
    mux.Handle("/api/v1/sources/", middleware.AuthMiddleware(te.Cfg.JWT_SECRET)(http.HandlerFunc(sh.GetSourceStatusOrDelete)))
    log := logger.New("INFO")
    return middleware.RequestLogger(log)(mux)
}

func TestR2_EnvMissing_UploadFails_NoKeyLeakInLogs(t *testing.T) {
    te, err := SetupTestEnv()
    if err != nil { t.Fatalf("setup failed: %v", err) }
    defer TeardownTestEnv(te)

    h := buildR2Mux(te, "")
    srv := httptest.NewServer(h)
    defer srv.Close()

    // Capture stdout during the request lifecycle
    origStdout := os.Stdout
    pr, pw, _ := os.Pipe()
    os.Stdout = pw

    // register & login
    reg := map[string]string{"email": "r2env@example.com", "password": "pass1234", "full_name": "User"}
    rb, _ := json.Marshal(reg)
    resReg, _ := http.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(rb))
    if resReg.StatusCode != http.StatusCreated { t.Fatalf("expected 201, got %d", resReg.StatusCode) }
    var tr tokenResp
    json.NewDecoder(resReg.Body).Decode(&tr)
    resReg.Body.Close()

    // create chatbot
    create := map[string]any{"name": "R2 Bot"}
    cbj, _ := json.Marshal(create)
    reqC, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
    reqC.Header.Set("Authorization", "Bearer "+tr.Token)
    reqC.Header.Set("Content-Type", "application/json")
    resC, _ := http.DefaultClient.Do(reqC)
    if resC.StatusCode != http.StatusCreated { t.Fatalf("expected 201, got %d", resC.StatusCode) }
    var bot chatbot
    json.NewDecoder(resC.Body).Decode(&bot)
    resC.Body.Close()

    // upload text source → should 500 due to empty bucket
    var body bytes.Buffer
    mw := multipart.NewWriter(&body)
    mw.WriteField("source_type", "text")
    mw.WriteField("text", "deneme metni")
    mw.Close()
    reqS, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/chatbots/"+bot.ID+"/sources", bytes.NewReader(body.Bytes()))
    reqS.Header.Set("Authorization", "Bearer "+tr.Token)
    reqS.Header.Set("Content-Type", mw.FormDataContentType())
    resS, _ := http.DefaultClient.Do(reqS)
    if resS.StatusCode != http.StatusInternalServerError { t.Fatalf("expected 500, got %d", resS.StatusCode) }
    resS.Body.Close()

    // Stop capture and read logs
    pw.Close()
    os.Stdout = origStdout
    out, _ := io.ReadAll(pr)
    pr.Close()

    // Assert secret-like values are not present in stdout logs
    if bytes.Contains(out, []byte("AKIA_TEST_LEAK")) || bytes.Contains(out, []byte("SECRET_TEST_LEAK")) {
        t.Fatalf("secret keys leaked in logs")
    }
}
