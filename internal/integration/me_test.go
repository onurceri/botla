package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestMe_ReturnsUserProfile(t *testing.T) {
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
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.ID == "" {
		t.Errorf("expected id to be set")
	}
	if body.Email != "me@example.com" {
		t.Errorf("expected email me@example.com, got %s", body.Email)
	}
}

func TestFreePlan_DefaultAssignment_RegisterThenMe(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := fmt.Sprintf("free-default-%d@example.com", time.Now().UnixNano())
	regBody := map[string]string{
		"email":     email,
		"password":  "Test@123",
		"full_name": "Free Default",
	}
	b, err := json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}

	var tr tokenResp
	if err = json.NewDecoder(res.Body).Decode(&tr); err != nil {
		t.Fatalf("decode register response failed: %v", err)
	}
	res.Body.Close()
	if tr.Token == "" {
		t.Fatalf("token empty")
	}

	// Verify plan assignment via /me/plan
	req, err := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me/plan", nil)
	if err != nil {
		t.Fatalf("failed to create /me/plan request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tr.Token)

	resMe, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to call /me/plan: %v", err)
	}
	if resMe.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /me/plan, got %d", resMe.StatusCode)
	}
	defer resMe.Body.Close()

	var body struct {
		Code string `json:"code"`
	}
	if err = json.NewDecoder(resMe.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode /me/plan response: %v", err)
	}

	if body.Code != "free" {
		t.Fatalf("expected plan code free, got %s", body.Code)
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

func TestMe_DBErrorReturns500(t *testing.T) {
	te, setupErr := SetupTestEnv()
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}
	defer TeardownTestEnv(te)

	closeErr := te.DB.Close()
	if closeErr != nil {
		t.Fatalf("failed to close db: %v", closeErr)
	}

	token := authToken(t, te.Server.URL, "me-dberror@example.com")
	req, err := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to call /me: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 when db is closed, got %d", res.StatusCode)
	}
}
