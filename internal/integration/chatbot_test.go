package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"
)

type chatbot struct {
    ID string `json:"id"`
    UserID string `json:"user_id"`
    Name string `json:"name"`
}

func authToken(t *testing.T, base string, email string) string {
    regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "User"}
    b, _ := json.Marshal(regBody)
    http.Post(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
    lb := map[string]string{"email": email, "password": "pass1234"}
    lbj, _ := json.Marshal(lb)
    res, _ := http.Post(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
    var tr tokenResp
    json.NewDecoder(res.Body).Decode(&tr)
    res.Body.Close()
    return tr.Token
}

func TestChatbot_CRUD(t *testing.T) {
    te, err := SetupTestEnv()
    if err != nil {
        t.Fatalf("setup failed: %v", err)
    }
    defer TeardownTestEnv(te)
    token := authToken(t, te.Server.URL, "crud@example.com")

    // create
    create := map[string]any{"name": "My Bot"}
    cb, _ := json.Marshal(create)
    req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    res, _ := http.DefaultClient.Do(req)
    if res.StatusCode != http.StatusCreated {
        t.Fatalf("expected 201, got %d", res.StatusCode)
    }
    var created chatbot
    json.NewDecoder(res.Body).Decode(&created)
    res.Body.Close()
    if created.ID == "" || created.Name != "My Bot" {
        t.Fatalf("invalid create response")
    }

    // list
    req2, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots", nil)
    req2.Header.Set("Authorization", "Bearer "+token)
    res2, _ := http.DefaultClient.Do(req2)
    if res2.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", res2.StatusCode)
    }
    var items []chatbot
    json.NewDecoder(res2.Body).Decode(&items)
    res2.Body.Close()
    if len(items) == 0 {
        t.Fatalf("expected at least 1 item")
    }

    // get by id
    req3, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+created.ID, nil)
    req3.Header.Set("Authorization", "Bearer "+token)
    res3, _ := http.DefaultClient.Do(req3)
    if res3.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", res3.StatusCode)
    }
    res3.Body.Close()

    // update
    upd := map[string]any{"name": "Renamed Bot"}
    ub, _ := json.Marshal(upd)
    req4, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+created.ID, bytes.NewReader(ub))
    req4.Header.Set("Authorization", "Bearer "+token)
    req4.Header.Set("Content-Type", "application/json")
    res4, _ := http.DefaultClient.Do(req4)
    if res4.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", res4.StatusCode)
    }
    var updated chatbot
    json.NewDecoder(res4.Body).Decode(&updated)
    res4.Body.Close()
    if updated.Name != "Renamed Bot" {
        t.Fatalf("name not updated")
    }

    // delete
    req5, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/chatbots/"+created.ID, nil)
    req5.Header.Set("Authorization", "Bearer "+token)
    res5, _ := http.DefaultClient.Do(req5)
    if res5.StatusCode != http.StatusNoContent {
        t.Fatalf("expected 204, got %d", res5.StatusCode)
    }
    res5.Body.Close()

    // get after delete
    req6, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+created.ID, nil)
    req6.Header.Set("Authorization", "Bearer "+token)
    res6, _ := http.DefaultClient.Do(req6)
    if res6.StatusCode != http.StatusNotFound {
        t.Fatalf("expected 404, got %d", res6.StatusCode)
    }
    res6.Body.Close()
}

