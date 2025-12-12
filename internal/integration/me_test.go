package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestMe_ReturnsSubscriptionPlan(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	token := authToken(t, te.Server.URL, "me@example.com")
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	defer res.Body.Close()

	var body struct {
		PlanCode string `json:"plan_code"`
		Config   struct {
			Chat struct {
				AllowedModels    []string `json:"allowed_models"`
				MaxMonthlyTokens int      `json:"max_monthly_tokens"`
			} `json:"chat"`
		} `json:"config"`
		Usage struct {
			FilesCount    int `json:"files_count"`
			StorageUsedMB int `json:"storage_used_mb"`
			URLsCount     int `json:"urls_count"`
			TokensUsed    int `json:"tokens_used"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.PlanCode != "free" {
		t.Errorf("expected plan code free, got %s", body.PlanCode)
	}
	if len(body.Config.Chat.AllowedModels) == 0 {
		t.Errorf("expected at least one allowed model")
	}
	if body.Config.Chat.MaxMonthlyTokens == 0 {
		t.Errorf("expected max monthly tokens to be set")
	}

	if body.Usage.FilesCount != 0 {
		t.Errorf("expected 0 files, got %d", body.Usage.FilesCount)
	}
	if body.Usage.URLsCount != 0 {
		t.Errorf("expected 0 urls, got %d", body.Usage.URLsCount)
	}
	if body.Usage.TokensUsed != 0 {
		t.Errorf("expected 0 tokens, got %d", body.Usage.TokensUsed)
	}
}

func TestMe_ProfileBasicInfo(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "me-basic@example.com"
	token := authToken(t, te.Server.URL, email)
	req, err := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to call /me: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	defer res.Body.Close()

	var body struct {
		ID        string  `json:"id"`
		Email     string  `json:"email"`
		FullName  *string `json:"full_name"`
		CreatedAt string  `json:"created_at"`
	}
	if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.ID == "" {
		t.Fatalf("expected non-empty id")
	}
	if body.Email != email {
		t.Fatalf("expected email %s, got %s", email, body.Email)
	}
	if body.FullName == nil || *body.FullName == "" {
		t.Fatalf("expected full_name to be set")
	}
	if body.CreatedAt == "" {
		t.Fatalf("expected created_at to be set")
	}
}

func TestMe_ProfileIncludesOrganizations(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "me-orgs@example.com"
	token := authToken(t, te.Server.URL, email)
	req, err := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to call /me: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	defer res.Body.Close()

	var body struct {
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Role string `json:"role"`
		} `json:"organizations"`
	}
	if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Organizations) == 0 {
		t.Fatalf("expected at least one organization")
	}
	foundOwner := false
	for _, org := range body.Organizations {
		if org.Role == "owner" {
			foundOwner = true
		}
		if org.ID == "" {
			t.Fatalf("expected organization id to be set")
		}
		if org.Name == "" {
			t.Fatalf("expected organization name to be set")
		}
	}
	if !foundOwner {
		t.Fatalf("expected at least one organization with owner role")
	}
}

func TestMe_CrossUserProfileIsolation(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	emailA := "user-a@example.com"
	emailB := "user-b@example.com"

	tokenA := authToken(t, te.Server.URL, emailA)
	tokenB := authToken(t, te.Server.URL, emailB)

	reqA, err := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	if err != nil {
		t.Fatalf("failed to create request A: %v", err)
	}
	reqA.Header.Set("Authorization", "Bearer "+tokenA)
	resA, err := http.DefaultClient.Do(reqA)
	if err != nil {
		t.Fatalf("failed to call /me for user A: %v", err)
	}
	if resA.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for user A, got %d", resA.StatusCode)
	}
	defer resA.Body.Close()

	var bodyA struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err = json.NewDecoder(resA.Body).Decode(&bodyA); err != nil {
		t.Fatalf("failed to decode response for user A: %v", err)
	}

	reqB, err := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	if err != nil {
		t.Fatalf("failed to create request B: %v", err)
	}
	reqB.Header.Set("Authorization", "Bearer "+tokenB)
	resB, err := http.DefaultClient.Do(reqB)
	if err != nil {
		t.Fatalf("failed to call /me for user B: %v", err)
	}
	if resB.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for user B, got %d", resB.StatusCode)
	}
	defer resB.Body.Close()

	var bodyB struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err = json.NewDecoder(resB.Body).Decode(&bodyB); err != nil {
		t.Fatalf("failed to decode response for user B: %v", err)
	}

	if bodyA.ID == "" || bodyB.ID == "" {
		t.Fatalf("expected non-empty ids for both users")
	}
	if bodyA.ID == bodyB.ID {
		t.Fatalf("expected different ids for different users")
	}
	if bodyA.Email != emailA {
		t.Fatalf("expected email %s for user A, got %s", emailA, bodyA.Email)
	}
	if bodyB.Email != emailB {
		t.Fatalf("expected email %s for user B, got %s", emailB, bodyB.Email)
	}
}
