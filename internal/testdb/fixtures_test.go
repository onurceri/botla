package testdb

import (
	"testing"

	"github.com/onurceri/botla-co/pkg/policy"
)

func TestCreateUser(t *testing.T) {
	db := OpenParallelTestDB(t)

	t.Run("creates user with defaults", func(t *testing.T) {
		user := CreateUser(t, db)

		if user == nil {
			t.Fatal("expected user to be created")
		}
		if user.ID == "" {
			t.Error("expected user ID to be set")
		}
		if user.Email == "" {
			t.Error("expected user email to be set")
		}
	})

	t.Run("creates user with custom email", func(t *testing.T) {
		user := CreateUser(t, db, UserFixture{
			Email: "custom@example.com",
		})

		if user.Email != "custom@example.com" {
			t.Errorf("expected email 'custom@example.com', got %q", user.Email)
		}
	})

	t.Run("creates platform admin user", func(t *testing.T) {
		user := CreateUser(t, db, UserFixture{
			IsPlatformAdmin: true,
		})

		if !user.IsPlatformAdmin {
			t.Error("expected user to be platform admin")
		}
	})

	t.Run("creates user with specific plan", func(t *testing.T) {
		user := CreateUser(t, db, UserFixture{
			PlanCode: policy.PlanPro.String(),
		})

		if user.PlanID == nil || *user.PlanID == "" {
			t.Error("expected plan ID to be set")
		}
	})
}

func TestCreateOrganization(t *testing.T) {
	db := OpenParallelTestDB(t)

	t.Run("creates organization with defaults", func(t *testing.T) {
		result := CreateOrganization(t, db)

		if result.Organization == nil {
			t.Fatal("expected organization to be created")
		}
		if result.Owner == nil {
			t.Fatal("expected owner to be created")
		}
		if result.Organization.OwnerID != result.Owner.ID {
			t.Error("expected organization owner to match created user")
		}
	})

	t.Run("creates organization with existing owner", func(t *testing.T) {
		existingUser := CreateUser(t, db)
		result := CreateOrganization(t, db, OrganizationFixture{
			OwnerID: existingUser.ID,
		})

		if result.Organization.OwnerID != existingUser.ID {
			t.Errorf("expected owner ID %q, got %q", existingUser.ID, result.Organization.OwnerID)
		}
	})

	t.Run("creates organization with custom name and slug", func(t *testing.T) {
		result := CreateOrganization(t, db, OrganizationFixture{
			Name: "Acme Corp",
			Slug: "acme-corp",
		})

		if result.Organization.Name != "Acme Corp" {
			t.Errorf("expected name 'Acme Corp', got %q", result.Organization.Name)
		}
		if result.Organization.Slug != "acme-corp" {
			t.Errorf("expected slug 'acme-corp', got %q", result.Organization.Slug)
		}
	})
}

func TestCreateWorkspace(t *testing.T) {
	db := OpenParallelTestDB(t)

	t.Run("creates workspace with full hierarchy", func(t *testing.T) {
		result := CreateWorkspace(t, db)

		if result.Workspace == nil {
			t.Fatal("expected workspace to be created")
		}
		if result.Organization == nil {
			t.Fatal("expected organization to be created")
		}
		if result.Owner == nil {
			t.Fatal("expected owner to be created")
		}
		if result.Workspace.OrganizationID != result.Organization.ID {
			t.Error("expected workspace organization to match")
		}
	})

	t.Run("creates workspace for existing organization", func(t *testing.T) {
		orgResult := CreateOrganization(t, db)
		wsResult := CreateWorkspace(t, db, WorkspaceFixture{
			OrganizationID: orgResult.Organization.ID,
			Name:           "Marketing Workspace",
			Slug:           "marketing-ws",
		})

		if wsResult.Workspace.OrganizationID != orgResult.Organization.ID {
			t.Errorf("expected org ID %q, got %q", orgResult.Organization.ID, wsResult.Workspace.OrganizationID)
		}
		if wsResult.Workspace.Name != "Marketing Workspace" {
			t.Errorf("expected name 'Marketing Workspace', got %q", wsResult.Workspace.Name)
		}
	})
}

func TestCreateChatbot(t *testing.T) {
	db := OpenParallelTestDB(t)

	t.Run("creates chatbot with full hierarchy", func(t *testing.T) {
		result := CreateChatbot(t, db)

		if result.Chatbot == nil {
			t.Fatal("expected chatbot to be created")
		}
		if result.Workspace == nil {
			t.Fatal("expected workspace to be created")
		}
		if result.Organization == nil {
			t.Fatal("expected organization to be created")
		}
		if result.User == nil {
			t.Fatal("expected user to be created")
		}
		if result.Chatbot.UserID != result.User.ID {
			t.Error("expected chatbot user to match created user")
		}
	})

	t.Run("creates chatbot with custom properties", func(t *testing.T) {
		result := CreateChatbot(t, db, ChatbotFixture{
			Name:           "Support Bot",
			Model:          policy.ModelGPT4o.String(),
			WelcomeMessage: "Welcome to support!",
		})

		if result.Chatbot.Name != "Support Bot" {
			t.Errorf("expected name 'Support Bot', got %q", result.Chatbot.Name)
		}
		if result.Chatbot.Model != "gpt-4o" {
			t.Errorf("expected model 'gpt-4o', got %q", result.Chatbot.Model)
		}
		if result.Chatbot.WelcomeMessage != "Welcome to support!" {
			t.Errorf("expected welcome message 'Welcome to support!', got %q", result.Chatbot.WelcomeMessage)
		}
	})

	t.Run("creates chatbot for existing workspace", func(t *testing.T) {
		wsResult := CreateWorkspace(t, db)
		cbResult := CreateChatbot(t, db, ChatbotFixture{
			WorkspaceID: &wsResult.Workspace.ID,
			UserID:      wsResult.Owner.ID,
		})

		if cbResult.Chatbot.WorkspaceID == nil || *cbResult.Chatbot.WorkspaceID != wsResult.Workspace.ID {
			t.Error("expected chatbot workspace to match")
		}
	})
}

func TestCreateSource(t *testing.T) {
	db := OpenParallelTestDB(t)

	t.Run("creates source with full hierarchy", func(t *testing.T) {
		result := CreateSource(t, db)

		if result.Source == nil {
			t.Fatal("expected source to be created")
		}
		if result.Chatbot == nil {
			t.Fatal("expected chatbot to be created")
		}
		if result.Source.ChatbotID != result.Chatbot.ID {
			t.Error("expected source chatbot to match")
		}
	})

	t.Run("creates URL source", func(t *testing.T) {
		url := "https://example.com"
		result := CreateSource(t, db, SourceFixture{
			SourceType: "url",
			SourceURL:  &url,
		})

		if result.Source.SourceType != "url" {
			t.Errorf("expected source type 'url', got %q", result.Source.SourceType)
		}
		if result.Source.SourceURL == nil || *result.Source.SourceURL != url {
			t.Error("expected source URL to be set")
		}
	})

	t.Run("creates source for existing chatbot", func(t *testing.T) {
		cbResult := CreateChatbot(t, db)
		srcResult := CreateSource(t, db, SourceFixture{
			ChatbotID: cbResult.Chatbot.ID,
		})

		if srcResult.Source.ChatbotID != cbResult.Chatbot.ID {
			t.Errorf("expected chatbot ID %q, got %q", cbResult.Chatbot.ID, srcResult.Source.ChatbotID)
		}
	})
}

func TestStringPtr(t *testing.T) {
	s := "test"
	ptr := StringPtr(s)

	if ptr == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *ptr != s {
		t.Errorf("expected %q, got %q", s, *ptr)
	}
}
