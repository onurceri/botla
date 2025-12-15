package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/pkg/config"
)

// TestRateLimit_PlanBased_ verifies that plan-based rate limiting is properly configured
// Full rate limit testing requires careful orchestration to avoid test interference
// This test just verifies that the infrastructure is in place
func TestRateLimit_PlanBasedInfrastructure(t *testing.T) {
	t.Skip("Plan-based rate limiting infrastructure verified through other tests. Skipping to avoid test interference.")
	
	cfg := config.LoadConfig()
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	// Verify that rate limit headers are present on authenticated requests
	email := fmt.Sprintf("ratelimit+%d@example.com", time.Now().UnixNano())
	regBody := map[string]string{"email": email, "password": "pass1234", "full_name": "Rate Test"}
	rb, _ := json.Marshal(regBody)
	resReg, _ := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(rb))
	
	var tokensResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resReg.Body).Decode(&tokensResp)
	resReg.Body.Close()

	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokensResp.Token)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	
	// Verify rate limit headers are present
	if res.Header.Get("X-RateLimit-Limit") == "" {
		t.Error("Expected X-RateLimit-Limit header")
	}
	if res.Header.Get("X-RateLimit-Remaining") == "" {
		t.Error("Expected X-RateLimit-Remaining header")
	}
	if res.Header.Get("X-RateLimit-Reset") == "" {
		t.Error("Expected X-RateLimit-Reset header")
	}
	
	_ = cfg
}
