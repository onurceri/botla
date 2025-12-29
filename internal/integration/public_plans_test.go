package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPublicPlans_GetAllPlans(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// No auth required for public endpoint
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/plans", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var plans []struct {
		Code     string `json:"code"`
		Price    float64 `json:"price"`
		Currency string  `json:"currency"`
		Limits   struct {
			MaxChatbots int `json:"max_chatbots"`
		} `json:"limits"`
		Features struct {
			Chat struct {
				AllowedModels []string `json:"allowed_models"`
			} `json:"chat"`
		} `json:"features"`
	}

	if err := json.NewDecoder(res.Body).Decode(&plans); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(plans) == 0 {
		t.Error("expected at least one plan")
	}

	// Verify expected plans exist
	planCodes := make(map[string]bool)
	for _, p := range plans {
		planCodes[p.Code] = true
	}

	if !planCodes["free"] {
		t.Error("expected 'free' plan in response")
	}
}

func TestPublicPlans_GetPlanByCode(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Test getting 'free' plan
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/plans/free", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var plan struct {
		Code     string `json:"code"`
		Price    float64 `json:"price"`
		Currency string  `json:"currency"`
		Limits   struct {
			MaxChatbots int `json:"max_chatbots"`
		} `json:"limits"`
		Features struct {
			Chat struct {
				AllowedModels []string `json:"allowed_models"`
			} `json:"chat"`
			Files struct {
				MaxFilesTotal int `json:"max_files_total"`
			} `json:"files"`
		} `json:"features"`
	}

	if err := json.NewDecoder(res.Body).Decode(&plan); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if plan.Code != "free" {
		t.Errorf("expected code 'free', got '%s'", plan.Code)
	}

	if plan.Limits.MaxChatbots == 0 {
		t.Error("expected max_chatbots to be set")
	}
}

func TestPublicPlans_GetPlanByCode_NotFound(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/plans/nonexistent-plan", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", res.StatusCode)
	}
}

func TestPublicPlans_NoAuthRequired(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Verify /api/v1/plans doesn't require auth
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/plans", nil)
	// Explicitly NOT setting Authorization header
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	// Should succeed without auth
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200 without auth, got %d", res.StatusCode)
	}
}
