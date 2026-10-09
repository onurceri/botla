package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/pkg/policy"
)

// setupHandoffBot registers a pro-plan user with a handoff-enabled chatbot and
// returns the auth token and chatbot ID.
func setupHandoffBot(t *testing.T, te *fixtures.TestEnv, email string) (string, string) {
	t.Helper()
	token := authTokenForHandoff(t, te.Server.URL, email)
	updateProPlanConfig(t, te)
	var proPlanID string
	if err := te.DB.QueryRow("SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID); err != nil {
		t.Fatalf("lookup pro plan: %v", err)
	}
	if _, err := te.DB.Exec("UPDATE users SET plan_id=$1 WHERE email=$2", proPlanID, email); err != nil {
		t.Fatalf("assign pro plan: %v", err)
	}

	cb, _ := json.Marshal(map[string]any{"name": "Handoff Bot", "language": "en-US"})
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, err := testHTTPClient().Do(reqC)
	if err != nil {
		t.Fatalf("create bot: %v", err)
	}
	defer drainBody(resC)
	if resC.StatusCode != http.StatusCreated {
		t.Fatalf("create bot failed: %d", resC.StatusCode)
	}
	var bot struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resC.Body).Decode(&bot); err != nil {
		t.Fatalf("decode bot: %v", err)
	}

	ub, _ := json.Marshal(map[string]any{
		"handoff_enabled": true,
		"handoff_type":    "email",
		"handoff_config":  map[string]string{"email_to": "support@example.com"},
	})
	reqU, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+bot.ID, bytes.NewReader(ub))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	resU, err := testHTTPClient().Do(reqU)
	if err != nil {
		t.Fatalf("enable handoff: %v", err)
	}
	defer drainBody(resU)
	if resU.StatusCode != http.StatusOK {
		t.Fatalf("enable handoff failed: %d", resU.StatusCode)
	}
	return token, bot.ID
}

// requestPublicHandoff opens a handoff for a fresh widget session and returns its ID.
func requestPublicHandoff(t *testing.T, te *fixtures.TestEnv, botID, sessionID string) string {
	t.Helper()
	hb, _ := json.Marshal(map[string]any{"session_id": sessionID, "message": "human please"})
	res, err := testHTTPPost(te.Server.URL+"/api/v1/public/chatbots/"+botID+"/handoff", "application/json", bytes.NewReader(hb))
	if err != nil {
		t.Fatalf("request handoff: %v", err)
	}
	defer drainBody(res)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("request handoff failed: %d", res.StatusCode)
	}
	var out struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil || out.RequestID == "" {
		t.Fatalf("decode handoff response: %v (%+v)", err, out)
	}
	return out.RequestID
}

func submitHandoffContact(t *testing.T, te *fixtures.TestEnv, botID, requestID, email string) int {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q}`, email)
	url := te.Server.URL + "/api/v1/public/chatbots/" + botID + "/handoff/" + requestID + "/contact"
	res, err := testHTTPPost(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("submit contact: %v", err)
	}
	defer drainBody(res)
	return res.StatusCode
}

func TestHandoff_PublicSubmitEmail_PersistsContact(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token, botID := setupHandoffBot(t, te, "handoff_contact@example.com")
	requestID := requestPublicHandoff(t, te, botID, "contact-session-1")

	if code := submitHandoffContact(t, te, botID, requestID, "not-an-email"); code != http.StatusBadRequest {
		t.Fatalf("invalid email: expected 400, got %d", code)
	}
	if code := submitHandoffContact(t, te, botID, requestID, "visitor@example.com"); code != http.StatusOK {
		t.Fatalf("valid email: expected 200, got %d", code)
	}

	// Owner sees the email and contacted_at; status is untouched.
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+botID+"/handoff-requests/"+requestID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := testHTTPClient().Do(req)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	defer drainBody(res)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("get detail failed: %d", res.StatusCode)
	}
	var detail struct {
		Request models.HandoffRequest `json:"request"`
	}
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		t.Fatalf("decode detail: %v", err)
	}
	if detail.Request.UserEmail == nil || *detail.Request.UserEmail != "visitor@example.com" {
		t.Fatalf("expected user_email to be persisted, got %v", detail.Request.UserEmail)
	}
	if detail.Request.ContactedAt == nil || detail.Request.ContactedAt.IsZero() {
		t.Fatalf("expected contacted_at to be set")
	}
	if detail.Request.Status != models.HandoffStatusPending {
		t.Fatalf("expected status pending, got %s", detail.Request.Status)
	}

	// A request ID that belongs to another chatbot is not writable through this bot.
	_, otherBotID := setupHandoffBot(t, te, "handoff_contact_other@example.com")
	if code := submitHandoffContact(t, te, otherBotID, requestID, "attacker@example.com"); code != http.StatusNotFound {
		t.Fatalf("cross-bot submit: expected 404, got %d", code)
	}

	// Resolved requests no longer accept contact details.
	if _, err := te.DB.Exec("UPDATE handoff_requests SET status='resolved' WHERE id=$1", requestID); err != nil {
		t.Fatalf("resolve request: %v", err)
	}
	if code := submitHandoffContact(t, te, botID, requestID, "late@example.com"); code != http.StatusConflict {
		t.Fatalf("resolved request: expected 409, got %d", code)
	}
}

func TestHandoff_RequestOfAnotherChatbot_IsNotVisible(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	_, victimBotID := setupHandoffBot(t, te, "handoff_victim@example.com")
	requestID := requestPublicHandoff(t, te, victimBotID, "victim-session")
	attackerToken, attackerBotID := setupHandoffBot(t, te, "handoff_attacker@example.com")

	base := te.Server.URL + "/api/v1/chatbots/" + attackerBotID + "/handoff-requests/" + requestID

	reqGet, _ := http.NewRequest(http.MethodGet, base, nil)
	reqGet.Header.Set("Authorization", "Bearer "+attackerToken)
	resGet, err := testHTTPClient().Do(reqGet)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	defer drainBody(resGet)
	if resGet.StatusCode != http.StatusNotFound {
		t.Fatalf("detail via another chatbot: expected 404, got %d", resGet.StatusCode)
	}

	reqPatch, _ := http.NewRequest(http.MethodPatch, base, strings.NewReader(`{"status":"resolved"}`))
	reqPatch.Header.Set("Authorization", "Bearer "+attackerToken)
	reqPatch.Header.Set("Content-Type", "application/json")
	resPatch, err := testHTTPClient().Do(reqPatch)
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	defer drainBody(resPatch)
	if resPatch.StatusCode != http.StatusNotFound {
		t.Fatalf("patch via another chatbot: expected 404, got %d", resPatch.StatusCode)
	}

	var status string
	if err := te.DB.QueryRow("SELECT status FROM handoff_requests WHERE id=$1", requestID).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != models.HandoffStatusPending {
		t.Fatalf("victim request was modified: status=%s", status)
	}
}
