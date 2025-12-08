package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
)

func TestAction_CRUD(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	token := authTokenForAction(t, te.Server.URL, "action_crud@example.com")

	// Create chatbot
	createBot := map[string]any{"name": "Action Bot", "language": "en-US"}
	cb, _ := json.Marshal(createBot)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := http.DefaultClient.Do(reqC)
	if resC.StatusCode != http.StatusCreated {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resC.Body)
		t.Fatalf("expected 201, got %d. Body: %s", resC.StatusCode, buf.String())
	}
	var bot struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 1. Create Action
	createAction := map[string]any{
		"name":        "Test Action",
		"description": "A test action",
		"action_type": "http",
		"config":      map[string]string{"url": "https://example.com"},
		"enabled":     true,
	}
	ca, _ := json.Marshal(createAction)
	reqA, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions", bytes.NewReader(ca))
	reqA.Header.Set("Authorization", "Bearer "+token)
	reqA.Header.Set("Content-Type", "application/json")
	resA, _ := http.DefaultClient.Do(reqA)
	if resA.StatusCode != http.StatusCreated {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resA.Body)
		t.Fatalf("create action: expected 201, got %d. Body: %s", resA.StatusCode, buf.String())
	}
	var action models.ChatbotAction
	json.NewDecoder(resA.Body).Decode(&action)
	resA.Body.Close()

	if action.ID == "" || action.Name != "Test Action" {
		t.Fatalf("invalid action created")
	}

	// 2. List Actions
	reqL, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions", nil)
	reqL.Header.Set("Authorization", "Bearer "+token)
	resL, _ := http.DefaultClient.Do(reqL)
	if resL.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resL.Body)
		t.Fatalf("list actions: expected 200, got %d. Body: %s", resL.StatusCode, buf.String())
	}
	var listResp struct {
		Actions []models.ChatbotAction `json:"actions"`
	}
	json.NewDecoder(resL.Body).Decode(&listResp)
	resL.Body.Close()
	if len(listResp.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(listResp.Actions))
	}

	// 3. Get Action
	reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions/"+action.ID, nil)
	reqG.Header.Set("Authorization", "Bearer "+token)
	resG, _ := http.DefaultClient.Do(reqG)
	if resG.StatusCode != http.StatusOK {
		t.Fatalf("get action: expected 200, got %d", resG.StatusCode)
	}
	var gotAction models.ChatbotAction
	json.NewDecoder(resG.Body).Decode(&gotAction)
	resG.Body.Close()
	if gotAction.ID != action.ID {
		t.Fatalf("got wrong action id")
	}

	// 4. Update Action
	updateAction := map[string]any{
		"name":    "Updated Action",
		"enabled": false,
	}
	ua, _ := json.Marshal(updateAction)
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions/"+action.ID, bytes.NewReader(ua))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, _ := http.DefaultClient.Do(reqU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("update action: expected 200, got %d", resU.StatusCode)
	}
	var updatedAction models.ChatbotAction
	json.NewDecoder(resU.Body).Decode(&updatedAction)
	resU.Body.Close()
	if updatedAction.Name != "Updated Action" || updatedAction.Enabled != false {
		t.Fatalf("action not updated correctly")
	}

	// 5. Delete Action
	reqD, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions/"+action.ID, nil)
	reqD.Header.Set("Authorization", "Bearer "+token)
	resD, _ := http.DefaultClient.Do(reqD)
	if resD.StatusCode != http.StatusNoContent {
		t.Fatalf("delete action: expected 204, got %d", resD.StatusCode)
	}
	resD.Body.Close()

	// 6. Get after delete
	reqG2, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/actions/"+action.ID, nil)
	reqG2.Header.Set("Authorization", "Bearer "+token)
	resG2, _ := http.DefaultClient.Do(reqG2)
	if resG2.StatusCode != http.StatusNotFound {
		t.Fatalf("get deleted action: expected 404, got %d", resG2.StatusCode)
	}
	resG2.Body.Close()
}

func authTokenForAction(t *testing.T, base string, email string) string {
	regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "User"}
	b, _ := json.Marshal(regBody)
	http.Post(base+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	lb := map[string]string{"email": email, "password": "pass1234"}
	lbj, _ := json.Marshal(lb)
	res, _ := http.Post(base+"/api/v1/auth/login", "application/json", bytes.NewReader(lbj))
	var tr struct {
		Token string `json:"token"`
	}
	json.NewDecoder(res.Body).Decode(&tr)
	res.Body.Close()
	return tr.Token
}
