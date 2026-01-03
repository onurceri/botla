package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/policy"
)

type chatbot struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

func authToken(t *testing.T, base string, email string) string {
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "User"}
	b, _ := json.Marshal(regBody)
	regRes, _ := testHTTPPost(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	drainBody(regRes)

	lb := map[string]string{"email": email, "password": "Test@123"}
	lbj, _ := json.Marshal(lb)
	res, _ := testHTTPPost(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	if res == nil {
		t.Fatalf("login response is nil")
	}
	var tr tokenResp
	json.NewDecoder(res.Body).Decode(&tr)
	drainBody(res)
	return tr.Token
}

func TestChatbot_CRUD(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	token := authToken(t, te.Server.URL, "crud@example.com")

	// create
	create := map[string]any{"name": "My Bot", "language": "en-US"}
	cb, _ := json.Marshal(create)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := testHTTPClient().Do(req)
	if res.StatusCode != http.StatusCreated {
		drainBody(res)
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var created chatbot
	json.NewDecoder(res.Body).Decode(&created)
	drainBody(res)
	if created.ID == "" || created.Name != "My Bot" {
		t.Fatalf("invalid create response")
	}

	// list
	req2, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	res2, _ := testHTTPClient().Do(req2)
	if res2.StatusCode != http.StatusOK {
		drainBody(res2)
		t.Fatalf("expected 200, got %d", res2.StatusCode)
	}
	var items []chatbot
	json.NewDecoder(res2.Body).Decode(&items)
	drainBody(res2)
	if len(items) == 0 {
		t.Fatalf("expected at least 1 item")
	}

	// get by id
	req3, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+created.ID, nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	res3, _ := testHTTPClient().Do(req3)
	if res3.StatusCode != http.StatusOK {
		drainBody(res3)
		t.Fatalf("expected 200, got %d", res3.StatusCode)
	}
	drainBody(res3)

	// update
	upd := map[string]any{"name": "Renamed Bot"}
	ub, _ := json.Marshal(upd)
	req4, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+created.ID, bytes.NewReader(ub))
	req4.Header.Set("Authorization", "Bearer "+token)
	req4.Header.Set("Content-Type", "application/json")
	res4, _ := testHTTPClient().Do(req4)
	if res4.StatusCode != http.StatusOK {
		drainBody(res4)
		t.Fatalf("expected 200, got %d", res4.StatusCode)
	}
	var updated chatbot
	json.NewDecoder(res4.Body).Decode(&updated)
	drainBody(res4)
	if updated.Name != "Renamed Bot" {
		t.Fatalf("name not updated")
	}

	// delete
	req5, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/chatbots/"+created.ID, nil)
	req5.Header.Set("Authorization", "Bearer "+token)
	res5, _ := testHTTPClient().Do(req5)
	if res5.StatusCode != http.StatusNoContent {
		drainBody(res5)
		t.Fatalf("expected 204, got %d", res5.StatusCode)
	}
	drainBody(res5)

	// get after delete
	req6, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+created.ID, nil)
	req6.Header.Set("Authorization", "Bearer "+token)
	res6, _ := testHTTPClient().Do(req6)
	if res6.StatusCode != http.StatusNotFound {
		drainBody(res6)
		t.Fatalf("expected 404, got %d", res6.StatusCode)
	}
	drainBody(res6)
}

func TestFreePlan_ModelRestriction_AllowsGpt4oMini(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	email := "free-model-allowed@example.com"
	token := authToken(t, te.Server.URL, email)

	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), email)

	create := map[string]any{"name": "Free Plan Bot", "model": "gpt-4o-mini"}
	cb, _ := json.Marshal(create)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := testHTTPClient().Do(req)
	if res.StatusCode != http.StatusCreated {
		drainBody(res)
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var created chatbot
	json.NewDecoder(res.Body).Decode(&created)
	drainBody(res)
	if created.ID == "" {
		t.Fatalf("expected non-empty chatbot id")
	}
}

func TestFreePlan_ModelRestriction_BlocksDisallowedModels(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	email := "free-model-block@example.com"
	token := authToken(t, te.Server.URL, email)

	_, _ = te.DB.Exec(`UPDATE users SET plan_id=(SELECT id FROM plans WHERE code=$1) WHERE email=$2`, policy.PlanFree.String(), email)

	create := map[string]any{"name": "Free Plan Bot", "model": "gpt-4o-mini"}
	cb, _ := json.Marshal(create)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := testHTTPClient().Do(req)
	if res.StatusCode != http.StatusCreated {
		drainBody(res)
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var created chatbot
	json.NewDecoder(res.Body).Decode(&created)
	drainBody(res)
	if created.ID == "" {
		t.Fatalf("expected non-empty chatbot id")
	}

	for _, model := range []string{"gpt-4o", "claude-3-5-sonnet"} {
		upd := map[string]any{"model": model}
		ub, _ := json.Marshal(upd)
		reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+created.ID, bytes.NewReader(ub))
		reqU.Header.Set("Authorization", "Bearer "+token)
		reqU.Header.Set("Content-Type", "application/json")
		resU, _ := testHTTPClient().Do(reqU)
		if resU.StatusCode != http.StatusForbidden {
			drainBody(resU)
			t.Fatalf("expected 403 for model %s, got %d", model, resU.StatusCode)
		}
		drainBody(resU)
	}
}
