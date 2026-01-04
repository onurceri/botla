package scenarios

import (
	"context"
	"testing"

	"github.com/onurceri/botla-app/internal/integration"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealServices_CompleteUserJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full journey test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	ctx := context.Background()

	t.Run("full chatbot creation and chat flow", func(t *testing.T) {
		user := testdb.CreateUser(t, env.DB)
		t.Logf("Created user: %s", user.Email)

		orgResult := testdb.CreateOrganization(t, env.DB, testdb.OrganizationFixture{
			OwnerID: user.ID,
		})
		org := orgResult.Organization
		t.Logf("Created organization: %s", org.Name)

		wsResult := testdb.CreateWorkspace(t, env.DB, testdb.WorkspaceFixture{
			OrganizationID: org.ID,
		})
		ws := wsResult.Workspace
		t.Logf("Created workspace: %s", ws.Name)

		cbResult := testdb.CreateChatbot(t, env.DB, testdb.ChatbotFixture{
			WorkspaceID: &ws.ID,
			UserID:      user.ID,
		})
		cb := cbResult.Chatbot
		t.Logf("Created chatbot: %s", cb.Name)

		srcResult := testdb.CreateSource(t, env.DB, testdb.SourceFixture{
			ChatbotID: cb.ID,
		})
		source := srcResult.Source
		var sourceURL string
		if source.SourceURL != nil {
			sourceURL = *source.SourceURL
		}
		t.Logf("Created source: %s", sourceURL)

		var count int
		err := env.DB.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM users WHERE id = $1
		`, user.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "User should be persisted")

		err = env.DB.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM organizations WHERE id = $1
		`, org.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Organization should be persisted")

		err = env.DB.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM workspaces WHERE id = $1
		`, ws.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Workspace should be persisted")

		err = env.DB.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM chatbots WHERE id = $1
		`, cb.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Chatbot should be persisted")

		err = env.DB.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM data_sources WHERE id = $1
		`, source.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Source should be persisted")
	})
}

func TestRealServices_MultiTenantIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-tenant test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	ctx := context.Background()

	t.Run("organizations isolated by owner", func(t *testing.T) {
		user1 := testdb.CreateUser(t, env.DB)
		user2 := testdb.CreateUser(t, env.DB)

		org1Result := testdb.CreateOrganization(t, env.DB, testdb.OrganizationFixture{
			OwnerID: user1.ID,
		})
		org1 := org1Result.Organization

		org2Result := testdb.CreateOrganization(t, env.DB, testdb.OrganizationFixture{
			OwnerID: user2.ID,
		})
		org2 := org2Result.Organization

		var ownerID string
		err := env.DB.QueryRowContext(ctx, `
			SELECT owner_id FROM organizations WHERE id = $1
		`, org1.ID).Scan(&ownerID)
		require.NoError(t, err)
		assert.Equal(t, user1.ID, ownerID, "Org 1 owner should be user1")

		err = env.DB.QueryRowContext(ctx, `
			SELECT owner_id FROM organizations WHERE id = $1
		`, org2.ID).Scan(&ownerID)
		require.NoError(t, err)
		assert.Equal(t, user2.ID, ownerID, "Org 2 owner should be user2")
	})

	t.Run("chatbots isolated by organization", func(t *testing.T) {
		user1 := testdb.CreateUser(t, env.DB)
		user2 := testdb.CreateUser(t, env.DB)

		org1Result := testdb.CreateOrganization(t, env.DB, testdb.OrganizationFixture{
			OwnerID: user1.ID,
		})
		org1 := org1Result.Organization

		org2Result := testdb.CreateOrganization(t, env.DB, testdb.OrganizationFixture{
			OwnerID: user2.ID,
		})
		org2 := org2Result.Organization

		ws1Result := testdb.CreateWorkspace(t, env.DB, testdb.WorkspaceFixture{
			OrganizationID: org1.ID,
		})
		ws1 := ws1Result.Workspace

		ws2Result := testdb.CreateWorkspace(t, env.DB, testdb.WorkspaceFixture{
			OrganizationID: org2.ID,
		})
		ws2 := ws2Result.Workspace

		cb1Result := testdb.CreateChatbot(t, env.DB, testdb.ChatbotFixture{
			WorkspaceID:    &ws1.ID,
			OrganizationID: &org1.ID,
			UserID:         user1.ID,
		})
		cb1 := cb1Result.Chatbot

		cb2Result := testdb.CreateChatbot(t, env.DB, testdb.ChatbotFixture{
			WorkspaceID:    &ws2.ID,
			OrganizationID: &org2.ID,
			UserID:         user2.ID,
		})
		cb2 := cb2Result.Chatbot

		var orgID string
		err := env.DB.QueryRowContext(ctx, `
			SELECT organization_id FROM chatbots WHERE id = $1
		`, cb1.ID).Scan(&orgID)
		require.NoError(t, err)
		assert.Equal(t, org1.ID, orgID, "Chatbot 1 should belong to org 1")

		err = env.DB.QueryRowContext(ctx, `
			SELECT organization_id FROM chatbots WHERE id = $1
		`, cb2.ID).Scan(&orgID)
		require.NoError(t, err)
		assert.Equal(t, org2.ID, orgID, "Chatbot 2 should belong to org 2")
	})
}

func TestRealServices_DatabaseConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping constraints test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	ctx := context.Background()

	t.Run("unique email constraint", func(t *testing.T) {
		email := "unique-test@example.com"
		user := testdb.CreateUser(t, env.DB, testdb.UserFixture{
			Email: email,
		})
		_ = user // acknowledge the user was created

		// Try to create another user with the same email using direct INSERT
		// This should fail due to the unique constraint
		fixture := testdb.DefaultUserFixture()
		fixture.Email = email // Use the same email
		// Use dbConn directly to get the error
		ctx := context.Background()
		var err error
		_, err = env.DB.ExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, plan_id, is_platform_admin, onboarding_completed, onboarding_step, onboarding_skipped)
			VALUES ($1, $2, $3, $4, $5, $6, true, 0, false)
		`, fixture.ID, fixture.Email, "hashed_password", fixture.FullName, "agency_starter", false)
		assert.Error(t, err, "Should not allow duplicate email")
	})

	t.Run("foreign key constraint", func(t *testing.T) {
		_ = testdb.CreateOrganization(t, env.DB)

		_, err := env.DB.ExecContext(ctx, `
			INSERT INTO workspaces (id, organization_id, name, slug, created_at)
			VALUES (gen_random_uuid(), $1, 'Test WS', 'test-ws', NOW())
		`, "00000000-0000-0000-0000-000000000000")

		assert.Error(t, err, "Should not allow workspace with non-existent organization")
	})
}
