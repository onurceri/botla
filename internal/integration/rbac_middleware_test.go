package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

// Helper to create org and users with roles
func setupRBACEnv(t *testing.T) (*fixtures.TestEnv, string, map[string]string, string) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Clean up potential leftovers
	te.DB.Exec("TRUNCATE TABLE organizations, memberships, workspaces RESTART IDENTITY CASCADE")

	// Tokens
	tokens := make(map[string]string)
	tokens["owner"] = authToken(t, te.Server.URL, "owner@example.com")
	tokens["admin"] = authToken(t, te.Server.URL, "admin@example.com")
	tokens["member"] = authToken(t, te.Server.URL, "member@example.com")
	tokens["outsider"] = authToken(t, te.Server.URL, "outsider@example.com")

	// Create Org (Owner)
	createOrg := map[string]string{"name": "RBAC Org", "slug": "rbac-org"}
	orgBody, _ := json.Marshal(createOrg)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/organizations", bytes.NewReader(orgBody))
	req.Header.Set("Authorization", "Bearer "+tokens["owner"])
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create org: %d", res.StatusCode)
	}
	var org struct {
		ID string `json:"id"`
	}
	json.NewDecoder(res.Body).Decode(&org)
	res.Body.Close()

	// Add Admin
	addMember(t, te, tokens["owner"], org.ID, "admin@example.com", "admin")
	// Add Member
	addMember(t, te, tokens["owner"], org.ID, "member@example.com", "member")

	// Create a Workspace for workspace tests
	wsID := createWorkspace(t, te, tokens["admin"], org.ID, "WS1")

	return te, org.ID, tokens, wsID
}

func addMember(t *testing.T, te *fixtures.TestEnv, token, orgID, email, role string) {
	body, _ := json.Marshal(map[string]string{"email": email, "role": role})
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/organizations/"+orgID+"/members", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("failed to add member %s: %d", email, res.StatusCode)
	}
	res.Body.Close()
}

func createWorkspace(t *testing.T, te *fixtures.TestEnv, token, orgID, name string) string {
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	body, _ := json.Marshal(map[string]string{"name": name, "slug": slug})
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/organizations/"+orgID+"/workspaces", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create workspace: %d", res.StatusCode)
	}
	var ws struct {
		ID string `json:"id"`
	}
	json.NewDecoder(res.Body).Decode(&ws)
	res.Body.Close()
	return ws.ID
}

func TestRBAC_Matrix(t *testing.T) {
	te, orgID, tokens, wsID := setupRBACEnv(t)
	defer fixtures.TeardownTestEnv(te)

	// Get a dummy member ID for member routes
	// We need the ID of 'member@example.com'
	// Use owner to list members
	reqGet, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/organizations/"+orgID+"/members", nil)
	reqGet.Header.Set("Authorization", "Bearer "+tokens["owner"])
	resGet, _ := http.DefaultClient.Do(reqGet)
	var membersResp struct {
		Members []struct {
			UserID string `json:"user_id"`
			User   struct {
				Email string `json:"email"`
			} `json:"user"`
		} `json:"members"`
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

	tests := []struct {
		name     string
		method   string
		path     string
		role     string
		wantCode int
	}{
		// RBAC-011: Get Org (member) -> OK
		{"GetOrg_Member", http.MethodGet, "/api/v1/organizations/" + orgID, "member", 200},
		// RBAC-004: Admin accesses member-required endpoint
		{"GetOrg_Admin", http.MethodGet, "/api/v1/organizations/" + orgID, "admin", 200},
		// RBAC-012: Patch Org (owner) -> OK
		{"PatchOrg_Owner", http.MethodPatch, "/api/v1/organizations/" + orgID, "owner", 200},
		{"PatchOrg_Admin", http.MethodPatch, "/api/v1/organizations/" + orgID, "admin", 403},
		{"PatchOrg_Member", http.MethodPatch, "/api/v1/organizations/" + orgID, "member", 403},
		// RBAC-013: Delete Org (owner) -> OK (Done last to avoid breaking others)

		// RBAC-014: Get Workspaces (member) -> OK
		{"GetWS_Member", http.MethodGet, "/api/v1/organizations/" + orgID + "/workspaces", "member", 200},

		// RBAC-015: Create Workspace (admin) -> Created
		{"CreateWS_Admin", http.MethodPost, "/api/v1/organizations/" + orgID + "/workspaces", "admin", 201},
		{"CreateWS_Member", http.MethodPost, "/api/v1/organizations/" + orgID + "/workspaces", "member", 403},

		// RBAC-016: Patch Workspace (admin) -> OK
		{"PatchWS_Admin", http.MethodPatch, "/api/v1/organizations/" + orgID + "/workspaces/" + wsID, "admin", 200},
		{"PatchWS_Member", http.MethodPatch, "/api/v1/organizations/" + orgID + "/workspaces/" + wsID, "member", 403},

		// RBAC-017: Delete Workspace (admin) -> OK (Create new one to delete)

		// RBAC-018: Get Members (member) -> OK
		{"GetMembers_Member", http.MethodGet, "/api/v1/organizations/" + orgID + "/members", "member", 200},

		// RBAC-019: Add Member (admin) -> Created (Use dummy email)

		// RBAC-020: Patch Member (admin) -> OK
		{"PatchMember_Admin", http.MethodPatch, "/api/v1/organizations/" + orgID + "/members/" + memberUserID, "admin", 200},
		{"PatchMember_Member", http.MethodPatch, "/api/v1/organizations/" + orgID + "/members/" + memberUserID, "member", 403},

		// RBAC-008: Outsider access
		{"GetOrg_Outsider", http.MethodGet, "/api/v1/organizations/" + orgID, "outsider", 403},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var body []byte
			if tc.method == http.MethodPost || tc.method == http.MethodPatch {
				if tc.name == "CreateWS_Admin" || tc.name == "CreateWS_Member" {
					body, _ = json.Marshal(map[string]string{"name": "New WS", "slug": "new-ws"})
				} else if tc.name == "PatchWS_Admin" || tc.name == "PatchWS_Member" {
					body, _ = json.Marshal(map[string]string{"name": "Updated WS"})
				} else if tc.name == "PatchOrg_Owner" || tc.name == "PatchOrg_Admin" || tc.name == "PatchOrg_Member" {
					body, _ = json.Marshal(map[string]string{"name": "Updated Org"})
				} else if tc.name == "PatchMember_Admin" || tc.name == "PatchMember_Member" {
					body, _ = json.Marshal(map[string]string{"role": "member"}) // No change actually
				}
			}

			req, _ := http.NewRequest(tc.method, te.Server.URL+tc.path, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+tokens[tc.role])
			if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			res, _ := http.DefaultClient.Do(req)
			res.Body.Close()

			if res.StatusCode != tc.wantCode {
				t.Errorf("%s: expected %d, got %d", tc.name, tc.wantCode, res.StatusCode)
			}
		})
	}
}

func TestRBAC_Extended(t *testing.T) {
	te, orgID, tokens, _ := setupRBACEnv(t)
	defer fixtures.TeardownTestEnv(te)

	// RBAC-009: Missing Authorization header
	t.Run("MissingAuthHeader", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/organizations/"+orgID, nil)
		// No Authorization header
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d", res.StatusCode)
		}
	})

	// RBAC-010: Invalid organization ID in path
	t.Run("InvalidOrgID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/organizations/invalid-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+tokens["owner"])
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", res.StatusCode)
			return
		}
		var body struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(res.Body).Decode(&body)
		if body.Error != "Invalid ID format" {
			t.Errorf("expected error %q, got %q", "Invalid ID format", body.Error)
		}
	})

	t.Run("InvalidWorkspaceID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/organizations/"+orgID+"/workspaces/invalid-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+tokens["admin"])
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", res.StatusCode)
			return
		}
		var body struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(res.Body).Decode(&body)
		if body.Error != "Invalid ID format" {
			t.Errorf("expected error %q, got %q", "Invalid ID format", body.Error)
		}
	})

	t.Run("InvalidChatbotID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/invalid-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+tokens["owner"])
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", res.StatusCode)
			return
		}
		var body struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(res.Body).Decode(&body)
		if body.Error != "Invalid ID format" {
			t.Errorf("expected error %q, got %q", "Invalid ID format", body.Error)
		}
	})

	// RBAC-017: DELETE workspace (admin)
	t.Run("DeleteWorkspace_Admin", func(t *testing.T) {
		// Create a temp workspace to delete
		tempWSID := createWorkspace(t, te, tokens["admin"], orgID, "Temp WS")

		req, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/organizations/"+orgID+"/workspaces/"+tempWSID, nil)
		req.Header.Set("Authorization", "Bearer "+tokens["admin"])
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
			t.Errorf("expected 200/204, got %d", res.StatusCode)
		}
	})

	// RBAC-019: Add Member (admin)
	t.Run("AddMember_Admin", func(t *testing.T) {
		// Register user first
		authToken(t, te.Server.URL, "newmember@example.com")
		body, _ := json.Marshal(map[string]string{"email": "newmember@example.com", "role": "member"})
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/organizations/"+orgID+"/members", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokens["admin"])
		req.Header.Set("Content-Type", "application/json")
		res, _ := http.DefaultClient.Do(req)
		res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 Created, got %d", res.StatusCode)
		}
	})

	// RBAC-021: Remove Member (admin)
	t.Run("RemoveMember_Admin", func(t *testing.T) {
		// First add a member to remove
		authToken(t, te.Server.URL, "toremove@example.com")
		addMember(t, te, tokens["owner"], orgID, "toremove@example.com", "member")

		// Get the user ID
		reqGet, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/organizations/"+orgID+"/members", nil)
		reqGet.Header.Set("Authorization", "Bearer "+tokens["owner"])
		resGet, _ := http.DefaultClient.Do(reqGet)
		var membersResp struct {
			Members []struct {
				UserID string `json:"user_id"`
				User   struct {
					Email string `json:"email"`
				} `json:"user"`
			} `json:"members"`
		}
		json.NewDecoder(resGet.Body).Decode(&membersResp)
		resGet.Body.Close()

		var removeUserID string
		for _, m := range membersResp.Members {
			if m.User.Email == "toremove@example.com" {
				removeUserID = m.UserID
				break
			}
		}

		if removeUserID == "" {
			t.Fatalf("failed to find member to remove")
		}

		req, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/organizations/"+orgID+"/members/"+removeUserID, nil)
		req.Header.Set("Authorization", "Bearer "+tokens["admin"])
		res, _ := http.DefaultClient.Do(req)
		res.Body.Close()
		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
			t.Errorf("expected 200/204, got %d", res.StatusCode)
		}
	})
}

func TestRBAC_DeleteOrg(t *testing.T) {
	te, orgID, tokens, _ := setupRBACEnv(t)
	defer fixtures.TeardownTestEnv(te)

	// Admin cannot delete org
	reqDel, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/organizations/"+orgID, nil)
	reqDel.Header.Set("Authorization", "Bearer "+tokens["admin"])
	resDel, _ := http.DefaultClient.Do(reqDel)
	if resDel.StatusCode != http.StatusForbidden {
		t.Errorf("Admin deleted org, expected 403, got %d", resDel.StatusCode)
	}
	resDel.Body.Close()

	// Owner can delete org
	reqDelOwner, _ := http.NewRequest(http.MethodDelete, te.Server.URL+"/api/v1/organizations/"+orgID, nil)
	reqDelOwner.Header.Set("Authorization", "Bearer "+tokens["owner"])
	resDelOwner, _ := http.DefaultClient.Do(reqDelOwner)
	if resDelOwner.StatusCode != http.StatusOK && resDelOwner.StatusCode != http.StatusNoContent {
		t.Errorf("Owner failed to delete org, expected 200/204, got %d", resDelOwner.StatusCode)
	}
	resDelOwner.Body.Close()
}
