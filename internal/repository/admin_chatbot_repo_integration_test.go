package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr is a helper to create a pointer from a value.
func ptr[T any](v T) *T {
	return &v
}

// newAdminChatbotRepo creates a new PostgresAdminChatbotRepo for testing.
func newAdminChatbotRepo(t *testing.T) repository.AdminChatbotRepository {
	db := testdb.OpenParallelTestDB(t)
	return repository.NewPostgresAdminChatbotRepo(db)
}

// TestAdminChatbotRepo_ListChatbots_EmptyFilters tests listing chatbots with no filters.
func TestAdminChatbotRepo_ListChatbots_EmptyFilters(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Customer Service Bot", UserID: user.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Sales Assistant", UserID: user.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Customer Support", UserID: user.ID})

	filter := repository.AdminChatbotFilter{}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, total, "should find 3 chatbots")
	assert.Len(t, chatbots, 3)
}

// TestAdminChatbotRepo_ListChatbots_NameFilter tests filtering by name.
func TestAdminChatbotRepo_ListChatbots_NameFilter(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Support Bot", UserID: user.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "sales bot", UserID: user.ID})

	nameFilter := "SUPPORT"
	filter := repository.AdminChatbotFilter{Name: &nameFilter}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, total, "should find 1 chatbot case-insensitively")
	assert.Len(t, chatbots, 1)
	assert.Equal(t, "Support Bot", chatbots[0].Name)
}

// TestAdminChatbotRepo_ListChatbots_OrganizationFilter tests filtering by organization.
func TestAdminChatbotRepo_ListChatbots_OrganizationFilter(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	org1 := testdb.CreateOrganization(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.OrganizationFixture{OwnerID: user.ID})
	org2 := testdb.CreateOrganization(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.OrganizationFixture{OwnerID: user.ID})

	ws1 := testdb.CreateWorkspace(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.WorkspaceFixture{OrganizationID: org1.Organization.ID})
	ws2 := testdb.CreateWorkspace(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.WorkspaceFixture{OrganizationID: org2.Organization.ID})

	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID, WorkspaceID: &ws1.Workspace.ID, OrganizationID: &org1.Organization.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 2", UserID: user.ID, WorkspaceID: &ws1.Workspace.ID, OrganizationID: &org1.Organization.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 3", UserID: user.ID, WorkspaceID: &ws2.Workspace.ID, OrganizationID: &org2.Organization.ID})

	filter := repository.AdminChatbotFilter{OrganizationID: &org1.Organization.ID}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "should find 2 chatbots in org1")
	assert.Len(t, chatbots, 2)
}

// TestAdminChatbotRepo_ListChatbots_OwnerFilter tests filtering by owner.
func TestAdminChatbotRepo_ListChatbots_OwnerFilter(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user1 := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	user2 := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())

	org1 := testdb.CreateOrganization(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.OrganizationFixture{OwnerID: user1.ID})
	org2 := testdb.CreateOrganization(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.OrganizationFixture{OwnerID: user2.ID})

	ws1 := testdb.CreateWorkspace(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.WorkspaceFixture{OrganizationID: org1.Organization.ID})
	ws2 := testdb.CreateWorkspace(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.WorkspaceFixture{OrganizationID: org2.Organization.ID})

	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "User1 Bot 1", UserID: user1.ID, WorkspaceID: &ws1.Workspace.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "User1 Bot 2", UserID: user1.ID, WorkspaceID: &ws1.Workspace.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "User2 Bot", UserID: user2.ID, WorkspaceID: &ws2.Workspace.ID})

	filter := repository.AdminChatbotFilter{OwnerID: &user1.ID}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "should find 2 chatbots owned by user1")
	assert.Len(t, chatbots, 2)
}

// TestAdminChatbotRepo_ListChatbots_AllFilters tests combining all filters.
func TestAdminChatbotRepo_ListChatbots_AllFilters(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	org := testdb.CreateOrganization(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.OrganizationFixture{OwnerID: user.ID})
	ws := testdb.CreateWorkspace(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.WorkspaceFixture{OrganizationID: org.Organization.ID})

	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Alpha Support Bot", UserID: user.ID, WorkspaceID: &ws.Workspace.ID, OrganizationID: &org.Organization.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Beta Support Bot", UserID: user.ID, WorkspaceID: &ws.Workspace.ID, OrganizationID: &org.Organization.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Gamma Sales Bot", UserID: user.ID, WorkspaceID: &ws.Workspace.ID, OrganizationID: &org.Organization.ID})

	nameFilter := "support"
	filter := repository.AdminChatbotFilter{
		Name:           &nameFilter,
		OrganizationID: &org.Organization.ID,
		OwnerID:        &user.ID,
	}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "should find 2 chatbots matching all filters")
	assert.Len(t, chatbots, 2)

	names := []string{chatbots[0].Name, chatbots[1].Name}
	assert.Contains(t, names, "Alpha Support Bot")
	assert.Contains(t, names, "Beta Support Bot")
	assert.NotContains(t, names, "Gamma Sales Bot")
}

// TestAdminChatbotRepo_ListChatbots_Pagination tests pagination behavior.
func TestAdminChatbotRepo_ListChatbots_Pagination(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())

	for i := 0; i < 5; i++ {
		name := "Bot " + uuid.NewString()[:8]
		_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: name, UserID: user.ID})
		time.Sleep(1 * time.Millisecond)
	}

	filter := repository.AdminChatbotFilter{}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 2, 0)

	require.NoError(t, err)
	assert.Equal(t, 5, total, "total count should include all 5 chatbots")
	assert.Len(t, chatbots, 2, "should return only 2 chatbots per page")

	chatbots2, total2, err2 := repo.ListChatbots(ctx, filter, 2, 2)

	require.NoError(t, err2)
	assert.Equal(t, 5, total2, "total count should be same")
	assert.Len(t, chatbots2, 2, "should return second page")

	assert.NotEqual(t, chatbots[0].ID, chatbots2[0].ID, "pages should have different chatbots")
}

// TestAdminChatbotRepo_ListChatbots_DefaultPagination tests default limit/offset handling.
func TestAdminChatbotRepo_ListChatbots_DefaultPagination(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 2", UserID: user.ID})

	filter := repository.AdminChatbotFilter{}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 20, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, chatbots, 2)
}

// TestAdminChatbotRepo_ListChatbots_NegativeLimit tests negative limit defaults to 20.
func TestAdminChatbotRepo_ListChatbots_NegativeLimit(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 2", UserID: user.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 3", UserID: user.ID})

	filter := repository.AdminChatbotFilter{}
	chatbots, total, err := repo.ListChatbots(ctx, filter, -5, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, total, "should count all chatbots")
	assert.Len(t, chatbots, 3, "should default to limit of 20")
}

// TestAdminChatbotRepo_ListChatbots_ZeroLimit tests zero limit defaults to 20.
func TestAdminChatbotRepo_ListChatbots_ZeroLimit(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID})
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 2", UserID: user.ID})

	filter := repository.AdminChatbotFilter{}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 0, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "should count all chatbots")
	assert.Len(t, chatbots, 2, "should default to limit of 20")
}

// TestAdminChatbotRepo_ListChatbots_NegativeOffset tests negative offset defaults to 0.
func TestAdminChatbotRepo_ListChatbots_NegativeOffset(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID})

	filter := repository.AdminChatbotFilter{}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, -10)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "validate pagination")
	assert.Equal(t, 0, total)
	assert.Nil(t, chatbots)
}

// TestAdminChatbotRepo_ListChatbots_SubqueryResults tests source and message count subqueries.
func TestAdminChatbotRepo_ListChatbots_SubqueryResults(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	bot := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Test Bot", UserID: user.ID})

	_ = testdb.CreateSource(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "ready"})
	_ = testdb.CreateSource(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "ready"})
	_ = testdb.CreateSource(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "pending"})

	convID := uuid.NewString()
	pool := repo.(*repository.PostgresAdminChatbotRepo).Pool()
	_, err := pool.ExecContext(ctx, `
		INSERT INTO conversations (id, chatbot_id, created_at)
		VALUES ($1, $2, NOW())
	`, convID, bot.Chatbot.ID)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		msgID := uuid.NewString()
		_, err := pool.ExecContext(ctx, `
			INSERT INTO messages (id, conversation_id, role, content, created_at)
			VALUES ($1, $2, $3, 'test', NOW())
		`, msgID, convID, []string{"user", "bot"}[i%2])
		require.NoError(t, err)
	}

	filter := repository.AdminChatbotFilter{}
	chatbots, _, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Len(t, chatbots, 1)

	resultBot := chatbots[0]
	assert.Equal(t, 3, resultBot.SourceCount, "should count all non-deleted sources")
	assert.Equal(t, 3, resultBot.MessageCount, "should count all messages")
}

// TestAdminChatbotRepo_ListChatbots_JoinsCorrect tests that joins return correct data.
func TestAdminChatbotRepo_ListChatbots_JoinsCorrect(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	org := testdb.CreateOrganization(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.OrganizationFixture{OwnerID: user.ID, Name: "Test Organization LLC"})
	ws := testdb.CreateWorkspace(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.WorkspaceFixture{OrganizationID: org.Organization.ID})

	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{
		Name:           "Joined Bot",
		UserID:         user.ID,
		WorkspaceID:    &ws.Workspace.ID,
		OrganizationID: &org.Organization.ID,
	})

	filter := repository.AdminChatbotFilter{}
	chatbots, _, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	require.Len(t, chatbots, 1)

	bot := chatbots[0]
	assert.Equal(t, user.ID, bot.OwnerID, "should join with users")
	assert.Equal(t, user.Email, bot.OwnerEmail, "should have owner email from users join")
	assert.NotNil(t, bot.OrganizationID, "should have organization ID")
	assert.Equal(t, org.Organization.ID, *bot.OrganizationID, "should join with organizations")
	assert.NotNil(t, bot.OrganizationName, "should have organization name")
	assert.Equal(t, "Test Organization LLC", *bot.OrganizationName, "should have correct org name")
}

// TestAdminChatbotRepo_ListChatbots_NullFields tests handling of nullable fields.
func TestAdminChatbotRepo_ListChatbots_NullFields(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	org := testdb.CreateOrganization(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.OrganizationFixture{OwnerID: user.ID})
	ws := testdb.CreateWorkspace(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.WorkspaceFixture{OrganizationID: org.Organization.ID})

	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{
		Name:           "Minimal Bot",
		UserID:         user.ID,
		WorkspaceID:    &ws.Workspace.ID,
		OrganizationID: &org.Organization.ID,
	})

	nameFilter := "Minimal"
	filter := repository.AdminChatbotFilter{Name: &nameFilter}
	chatbots, _, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	require.Len(t, chatbots, 1)

	bot := chatbots[0]
	assert.NotNil(t, bot.WorkspaceID, "workspace_id should be set")
	assert.NotNil(t, bot.OrganizationID, "organization_id should be set")
	assert.NotNil(t, bot.OrganizationName, "organization_name should be set")
}

// TestAdminChatbotRepo_ListChatbots_SQLInjectionPrevention tests SQL injection prevention.
func TestAdminChatbotRepo_ListChatbots_SQLInjectionPrevention(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Normal Bot", UserID: user.ID})

	nameFilter := "' OR '1'='1"
	filter := repository.AdminChatbotFilter{Name: &nameFilter}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 0, total, "SQL injection attempt should not match anything")
	assert.Len(t, chatbots, 0, "should return no results for SQL injection attempt")
}

// TestAdminChatbotRepo_ListChatbots_SpecialCharacters tests handling of special characters.
func TestAdminChatbotRepo_ListChatbots_SpecialCharacters(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	_ = testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Bot With 'Quotes' & Symbols", UserID: user.ID})

	nameFilter := "'Quotes' & Symbols"
	filter := repository.AdminChatbotFilter{Name: &nameFilter}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, total, "should find bot with special characters in name")
	assert.Len(t, chatbots, 1)
	assert.Equal(t, "Bot With 'Quotes' & Symbols", chatbots[0].Name)
}

// TestAdminChatbotRepo_ListChatbots_EmptyResult tests empty result handling.
func TestAdminChatbotRepo_ListChatbots_EmptyResult(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	nameFilter := "NonExistentBot"
	filter := repository.AdminChatbotFilter{Name: &nameFilter}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 0, total, "should return 0 total for no matches")
	assert.NotNil(t, chatbots, "should return empty slice, not nil")
	assert.Len(t, chatbots, 0, "should return empty slice")
}

// TestAdminChatbotRepo_ListChatbots_DeletedExcluded tests that deleted chatbots are excluded.
func TestAdminChatbotRepo_ListChatbots_DeletedExcluded(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	bot1 := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Active Bot", UserID: user.ID})
	bot2 := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Deleted Bot", UserID: user.ID})

	pool := repo.(*repository.PostgresAdminChatbotRepo).Pool()
	_, err := pool.ExecContext(ctx, `UPDATE chatbots SET deleted_at = NOW() WHERE id = $1`, bot2.Chatbot.ID)
	require.NoError(t, err, "failed to soft delete bot")

	filter := repository.AdminChatbotFilter{}
	chatbots, total, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, total, "should only count non-deleted chatbots")
	assert.Len(t, chatbots, 1)
	assert.Equal(t, bot1.Chatbot.ID, chatbots[0].ID, "should return only active bot")
}

// TestAdminChatbotRepo_ListChatbots_SortByCreatedAtDesc tests sorting by created_at descending.
func TestAdminChatbotRepo_ListChatbots_SortByCreatedAtDesc(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())

	bot1 := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "First Bot", UserID: user.ID})
	time.Sleep(5 * time.Millisecond)
	bot2 := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Second Bot", UserID: user.ID})
	time.Sleep(5 * time.Millisecond)
	bot3 := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Third Bot", UserID: user.ID})

	filter := repository.AdminChatbotFilter{}
	chatbots, _, err := repo.ListChatbots(ctx, filter, 10, 0)

	require.NoError(t, err)
	require.Len(t, chatbots, 3)

	assert.Equal(t, bot3.Chatbot.ID, chatbots[0].ID, "first result should be newest (bot3)")
	assert.Equal(t, bot2.Chatbot.ID, chatbots[1].ID, "second result should be middle (bot2)")
	assert.Equal(t, bot1.Chatbot.ID, chatbots[2].ID, "third result should be oldest (bot1)")
}

// TestAdminChatbotRepo_GetByID_Success tests getting a single chatbot by ID.
func TestAdminChatbotRepo_GetByID_Success(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	// CreateChatbot creates its own hierarchy (workspace, org, user)
	bot := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Test Bot"})

	result, err := repo.GetByID(ctx, bot.Chatbot.ID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, bot.Chatbot.ID, result.ID, "ID should match")
	assert.Equal(t, "Test Bot", result.Name)
	assert.Equal(t, bot.User.ID, result.OwnerID)
}

// TestAdminChatbotRepo_GetByID_NotFound tests handling of non-existent chatbot.
func TestAdminChatbotRepo_GetByID_NotFound(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	// Use a valid UUID format that doesn't exist in the database
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	result, err := repo.GetByID(ctx, nonExistentID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, sql.ErrNoRows, err)
}

// TestAdminChatbotRepo_GetByID_DeletedExcluded tests that deleted chatbots are not returned.
func TestAdminChatbotRepo_GetByID_DeletedExcluded(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	bot := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Deleted Bot", UserID: user.ID})

	pool := repo.(*repository.PostgresAdminChatbotRepo).Pool()
	_, err := pool.ExecContext(ctx, `UPDATE chatbots SET deleted_at = NOW() WHERE id = $1`, bot.Chatbot.ID)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, bot.Chatbot.ID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, sql.ErrNoRows, err)
}

// TestAdminChatbotRepo_ResetSources_Success tests resetting sources for a chatbot.
func TestAdminChatbotRepo_ResetSources_Success(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	bot := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Test Bot", UserID: user.ID})
	pool := repo.(*repository.PostgresAdminChatbotRepo).Pool()

	// Create sources with different statuses
	_ = testdb.CreateSource(t, pool, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "ready"})
	_ = testdb.CreateSource(t, pool, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "failed"})
	_ = testdb.CreateSource(t, pool, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "pending"}) // Should not be reset

	count, err := repo.ResetSources(ctx, bot.Chatbot.ID)

	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "should reset 2 sources (ready and failed)")

	// Verify sources are now pending
	var status string
	err = pool.QueryRowContext(ctx, `SELECT status FROM data_sources WHERE chatbot_id = $1 AND status = 'ready'`, bot.Chatbot.ID).Scan(&status)
	assert.Error(t, err, "ready sources should be reset")
}

// TestAdminChatbotRepo_ResetSources_Empty tests resetting with no matching sources.
func TestAdminChatbotRepo_ResetSources_Empty(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	bot := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Test Bot", UserID: user.ID})

	count, err := repo.ResetSources(ctx, bot.Chatbot.ID)

	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "should reset 0 sources")
}

// TestAdminChatbotRepo_GetSourceIDs_Success tests getting pending source IDs.
func TestAdminChatbotRepo_GetSourceIDs_Success(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	bot := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Test Bot", UserID: user.ID})
	pool := repo.(*repository.PostgresAdminChatbotRepo).Pool()

	// Create sources with different statuses
	source1 := testdb.CreateSource(t, pool, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "pending"})
	source2 := testdb.CreateSource(t, pool, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "pending"})
	_ = testdb.CreateSource(t, pool, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "ready"})

	ids, err := repo.GetSourceIDs(ctx, bot.Chatbot.ID)

	require.NoError(t, err)
	assert.Len(t, ids, 2, "should return 2 pending source IDs")
	assert.Contains(t, ids, source1.Source.ID)
	assert.Contains(t, ids, source2.Source.ID)
}

// TestAdminChatbotRepo_GetSourceIDs_Empty tests getting source IDs when none are pending.
func TestAdminChatbotRepo_GetSourceIDs_Empty(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	bot := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Test Bot", UserID: user.ID})

	ids, err := repo.GetSourceIDs(ctx, bot.Chatbot.ID)

	require.NoError(t, err)
	assert.Nil(t, ids, "should return nil when no pending sources")
}

// TestAdminChatbotRepo_DeleteVectors_Success tests resetting chunk counts.
func TestAdminChatbotRepo_DeleteVectors_Success(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	bot := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Test Bot", UserID: user.ID})
	pool := repo.(*repository.PostgresAdminChatbotRepo).Pool()

	// Create sources with chunk counts
	source1 := testdb.CreateSource(t, pool, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "ready", ChunkCount: 10})
	source2 := testdb.CreateSource(t, pool, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "ready", ChunkCount: 5})

	err := repo.DeleteVectors(ctx, bot.Chatbot.ID)

	require.NoError(t, err)

	// Verify chunk counts are reset
	var count1, count2 int
	err = pool.QueryRowContext(ctx, `SELECT chunk_count FROM data_sources WHERE id = $1`, source1.Source.ID).Scan(&count1)
	require.NoError(t, err)
	assert.Equal(t, 0, count1, "chunk_count should be reset to 0")

	err = pool.QueryRowContext(ctx, `SELECT chunk_count FROM data_sources WHERE id = $1`, source2.Source.ID).Scan(&count2)
	require.NoError(t, err)
	assert.Equal(t, 0, count2, "chunk_count should be reset to 0")
}

// TestAdminChatbotRepo_DeleteVectors_Empty tests deleting vectors when no sources exist.
func TestAdminChatbotRepo_DeleteVectors_Empty(t *testing.T) {
	repo := newAdminChatbotRepo(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, repo.(*repository.PostgresAdminChatbotRepo).Pool())
	bot := testdb.CreateChatbot(t, repo.(*repository.PostgresAdminChatbotRepo).Pool(), testdb.ChatbotFixture{Name: "Test Bot", UserID: user.ID})

	err := repo.DeleteVectors(ctx, bot.Chatbot.ID)

	require.NoError(t, err)
	// No assertion needed - just verifying no error
}
