package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

type tokenRespSideEffect struct {
	Token string `json:"token"`
}

func TestAuth_RegistrationCreatesDefaultOrgAndWorkspace(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	email := "sideeffect+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	regBody := map[string]string{"email": email, "password": "Test@123", "full_name": "Side Effect User"}
	b, _ := json.Marshal(regBody)
	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}

	// Decode token to use for API calls
	var tr tokenRespSideEffect
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}
	res.Body.Close()

	// Check Organizations
	reqOrg, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/organizations", nil)
	reqOrg.Header.Set("Authorization", "Bearer "+tr.Token)
	resOrg, _ := http.DefaultClient.Do(reqOrg)
	if resOrg.StatusCode != http.StatusOK {
		t.Fatalf("failed to list orgs: %d", resOrg.StatusCode)
	}

	var orgs []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(resOrg.Body).Decode(&orgs); err != nil {
		t.Fatalf("failed to decode orgs: %v", err)
	}
	resOrg.Body.Close()

	if len(orgs) == 0 {
		t.Fatalf("expected at least 1 organization, got 0")
	}

	defaultOrg := orgs[0]
	// Expecting "Side Effect User Organizasyonu" or similar, or "Kişisel Organizasyon"
	// The requirement says: "<user_name> Organizasyonu" or "Kişisel Organizasyon"
	// We sent "Side Effect User" as full_name.

	if defaultOrg.Role != "owner" {
		t.Errorf("expected role owner, got %s", defaultOrg.Role)
	}

	// Check Workspaces for the default org
	reqWS, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/organizations/"+defaultOrg.ID+"/workspaces", nil)
	reqWS.Header.Set("Authorization", "Bearer "+tr.Token)
	resWS, _ := http.DefaultClient.Do(reqWS)
	if resWS.StatusCode != http.StatusOK {
		t.Fatalf("failed to list workspaces: %d", resWS.StatusCode)
	}

	var workspaces []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resWS.Body).Decode(&workspaces); err != nil {
		t.Fatalf("failed to decode workspaces: %v", err)
	}
	resWS.Body.Close()

	if len(workspaces) == 0 {
		t.Fatalf("expected at least 1 workspace, got 0")
	}

	foundDefault := false
	for _, ws := range workspaces {
		if ws.Name == "Varsayılan" || ws.Name == "Default" {
			foundDefault = true
			break
		}
	}

	if !foundDefault {
		t.Errorf("expected default workspace named 'Varsayılan' or 'Default', got names: %v", workspaces)
	}
}
