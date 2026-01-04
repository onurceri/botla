package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
)

func TestOrganization_UpdateMemberRole_Permissions(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// Create owner user and token
	ownerToken := authToken(t, te.Server.URL, "owner@example.com")

	// Create another user (to be member)
	memberToken := authToken(t, te.Server.URL, "member@example.com")

	// Create organization as owner
	createOrg := map[string]string{"name": "Test Org", "slug": "test-org"}
	orgBody, _ := json.Marshal(createOrg)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/organizations", bytes.NewReader(orgBody))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	res, _ := testHTTPClient().Do(req)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create org: %d", res.StatusCode)
	}
	var org struct {
		ID string `json:"id"`
	}
	json.NewDecoder(res.Body).Decode(&org)
	drainBody(res)

	// Add member to organization
	// First we need the member's ID. We can get it from the token logic or just add by email since AddMember looks up by email.
	// We'll use AddMember endpoint as owner
	addMember := map[string]string{"email": "member@example.com", "role": "member"}
	addBody, _ := json.Marshal(addMember)
	reqAdd, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/organizations/"+org.ID+"/members", bytes.NewReader(addBody))
	reqAdd.Header.Set("Authorization", "Bearer "+ownerToken)
	reqAdd.Header.Set("Content-Type", "application/json")
	resAdd, _ := testHTTPClient().Do(reqAdd)
	if resAdd.StatusCode != http.StatusCreated {
		t.Fatalf("failed to add member: %d", resAdd.StatusCode)
	}
	resAdd.Body.Close()

	// Get member's ID
	reqGet, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/organizations/"+org.ID+"/members", nil)
	reqGet.Header.Set("Authorization", "Bearer "+ownerToken)
	resGet, _ := testHTTPClient().Do(reqGet)

	// New response structure: { members: [...], caller_role: "..." }
	var membersResp struct {
		Members []struct {
			UserID string `json:"user_id"`
			User   struct {
				Email string `json:"email"`
			} `json:"user"`
			Role string `json:"role"`
		} `json:"members"`
		CallerRole string `json:"caller_role"`
	}
	json.NewDecoder(resGet.Body).Decode(&membersResp)
	resGet.Body.Close()

	var memberUserID string
	for _, m := range membersResp.Members {
		if m.User.Email == "member@example.com" {
			memberUserID = m.UserID
			break
		}
	}
	if memberUserID == "" {
		t.Fatalf("member not found in list")
	}

	// Case 1: Member tries to update their own role (Should fail)
	updateRole := map[string]string{"role": "admin"}
	updateBody, _ := json.Marshal(updateRole)
	reqUpdate, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/organizations/"+org.ID+"/members/"+memberUserID, bytes.NewReader(updateBody))
	reqUpdate.Header.Set("Authorization", "Bearer "+memberToken)
	reqUpdate.Header.Set("Content-Type", "application/json")
	resUpdate, _ := testHTTPClient().Do(reqUpdate)

	if resUpdate.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for member updating role, got %d", resUpdate.StatusCode)
	}
	resUpdate.Body.Close()

	// Case 2: Owner updates member's role (Should succeed)
	reqUpdateOwner, _ := http.NewRequest(http.MethodPatch, te.Server.URL+"/api/v1/organizations/"+org.ID+"/members/"+memberUserID, bytes.NewReader(updateBody))
	reqUpdateOwner.Header.Set("Authorization", "Bearer "+ownerToken)
	reqUpdateOwner.Header.Set("Content-Type", "application/json")
	resUpdateOwner, _ := testHTTPClient().Do(reqUpdateOwner)

	if resUpdateOwner.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for owner updating role, got %d", resUpdateOwner.StatusCode)
	}
	resUpdateOwner.Body.Close()
}
