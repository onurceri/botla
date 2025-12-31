package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr is a helper to create a pointer from a value
func ptr[T any](v T) *T {
	return &v
}

func TestAdminListChatbots_EmptyFilters(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Customer Service Bot", UserID: user.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Sales Assistant", UserID: user.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Customer Support", UserID: user.ID})

	nameFilter := "customer"
	filter := db.ChatbotFilter{Name: &nameFilter}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "should find 2 chatbots with 'customer' in name")
	assert.Len(t, chatbots, 2)

	names := []string{chatbots[0].Name, chatbots[1].Name}
	assert.Contains(t, names, "Customer Service Bot")
	assert.Contains(t, names, "Customer Support")
	assert.NotContains(t, names, "Sales Assistant")
}

func TestAdminListChatbots_NameFilterCaseInsensitive(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Support Bot", UserID: user.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "sales bot", UserID: user.ID})

	nameFilter := "SUPPORT"
	filter := db.ChatbotFilter{Name: &nameFilter}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, total, "should find 1 chatbot case-insensitively")
	assert.Len(t, chatbots, 1)
	assert.Equal(t, "Support Bot", chatbots[0].Name)
}

func TestAdminListChatbots_OrganizationFilter(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	org1 := testdb.CreateOrganization(t, dbConn, testdb.OrganizationFixture{OwnerID: user.ID})
	org2 := testdb.CreateOrganization(t, dbConn, testdb.OrganizationFixture{OwnerID: user.ID})

	ws1 := testdb.CreateWorkspace(t, dbConn, testdb.WorkspaceFixture{OrganizationID: org1.Organization.ID})
	ws2 := testdb.CreateWorkspace(t, dbConn, testdb.WorkspaceFixture{OrganizationID: org2.Organization.ID})

	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID, WorkspaceID: &ws1.Workspace.ID, OrganizationID: &org1.Organization.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 2", UserID: user.ID, WorkspaceID: &ws1.Workspace.ID, OrganizationID: &org1.Organization.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 3", UserID: user.ID, WorkspaceID: &ws2.Workspace.ID, OrganizationID: &org2.Organization.ID})

	filter := db.ChatbotFilter{OrganizationID: &org1.Organization.ID}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "should find 2 chatbots in org1")
	assert.Len(t, chatbots, 2)
}

func TestAdminListChatbots_OwnerFilter(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user1 := testdb.CreateUser(t, dbConn)
	user2 := testdb.CreateUser(t, dbConn)

	org1 := testdb.CreateOrganization(t, dbConn, testdb.OrganizationFixture{OwnerID: user1.ID})
	org2 := testdb.CreateOrganization(t, dbConn, testdb.OrganizationFixture{OwnerID: user2.ID})

	ws1 := testdb.CreateWorkspace(t, dbConn, testdb.WorkspaceFixture{OrganizationID: org1.Organization.ID})
	ws2 := testdb.CreateWorkspace(t, dbConn, testdb.WorkspaceFixture{OrganizationID: org2.Organization.ID})

	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "User1 Bot 1", UserID: user1.ID, WorkspaceID: &ws1.Workspace.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "User1 Bot 2", UserID: user1.ID, WorkspaceID: &ws1.Workspace.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "User2 Bot", UserID: user2.ID, WorkspaceID: &ws2.Workspace.ID})

	filter := db.ChatbotFilter{OwnerID: &user1.ID}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "should find 2 chatbots owned by user1")
	assert.Len(t, chatbots, 2)
}

func TestAdminListChatbots_AllFilters(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	org := testdb.CreateOrganization(t, dbConn, testdb.OrganizationFixture{OwnerID: user.ID})

	ws := testdb.CreateWorkspace(t, dbConn, testdb.WorkspaceFixture{OrganizationID: org.Organization.ID})

	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Alpha Support Bot", UserID: user.ID, WorkspaceID: &ws.Workspace.ID, OrganizationID: &org.Organization.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Beta Support Bot", UserID: user.ID, WorkspaceID: &ws.Workspace.ID, OrganizationID: &org.Organization.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Gamma Sales Bot", UserID: user.ID, WorkspaceID: &ws.Workspace.ID, OrganizationID: &org.Organization.ID})

	nameFilter := "support"
	filter := db.ChatbotFilter{
		Name:           &nameFilter,
		OrganizationID: &org.Organization.ID,
		OwnerID:        &user.ID,
	}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "should find 2 chatbots matching all filters")
	assert.Len(t, chatbots, 2)

	names := []string{chatbots[0].Name, chatbots[1].Name}
	assert.Contains(t, names, "Alpha Support Bot")
	assert.Contains(t, names, "Beta Support Bot")
	assert.NotContains(t, names, "Gamma Sales Bot")
}

func TestAdminListChatbots_Pagination(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)

	for i := 0; i < 5; i++ {
		name := "Bot " + uuid.NewString()[:8]
		_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: name, UserID: user.ID})
		time.Sleep(1 * time.Millisecond)
	}

	filter := db.ChatbotFilter{}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 2, 0)

	require.NoError(t, err)
	assert.Equal(t, 5, total, "total count should include all 5 chatbots")
	assert.Len(t, chatbots, 2, "should return only 2 chatbots per page")

	chatbots2, total2, err2 := db.AdminListChatbots(ctx, dbConn, filter, 2, 2)

	require.NoError(t, err2)
	assert.Equal(t, 5, total2, "total count should be same")
	assert.Len(t, chatbots2, 2, "should return second page")

	assert.NotEqual(t, chatbots[0].ID, chatbots2[0].ID, "pages should have different chatbots")
}

func TestAdminListChatbots_DefaultPagination(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)

	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 2", UserID: user.ID})

	filter := db.ChatbotFilter{}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 20, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, chatbots, 2)
}

func TestAdminListChatbots_SubqueryResults(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	bot := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Test Bot", UserID: user.ID})

	_ = testdb.CreateSource(t, dbConn, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "ready"})
	_ = testdb.CreateSource(t, dbConn, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "ready"})
	_ = testdb.CreateSource(t, dbConn, testdb.SourceFixture{ChatbotID: bot.Chatbot.ID, Status: "pending"})

	convID := uuid.NewString()
	_, err := dbConn.ExecContext(ctx, `
		INSERT INTO conversations (id, chatbot_id, created_at)
		VALUES ($1, $2, NOW())
	`, convID, bot.Chatbot.ID)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		msgID := uuid.NewString()
		_, err := dbConn.ExecContext(ctx, `
			INSERT INTO messages (id, conversation_id, role, content, created_at)
			VALUES ($1, $2, $3, 'test', NOW())
		`, msgID, convID, []string{"user", "bot"}[i%2])
		require.NoError(t, err)
	}

	filter := db.ChatbotFilter{}
	chatbots, _, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Len(t, chatbots, 1)

	resultBot := chatbots[0]
	assert.Equal(t, 3, resultBot.SourceCount, "should count all non-deleted sources")
	assert.Equal(t, 3, resultBot.MessageCount, "should count all messages")
}

func TestAdminListChatbots_JoinsCorrect(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	org := testdb.CreateOrganization(t, dbConn, testdb.OrganizationFixture{OwnerID: user.ID, Name: "Test Organization LLC"})
	ws := testdb.CreateWorkspace(t, dbConn, testdb.WorkspaceFixture{OrganizationID: org.Organization.ID})

	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{
		Name:           "Joined Bot",
		UserID:         user.ID,
		WorkspaceID:    &ws.Workspace.ID,
		OrganizationID: &org.Organization.ID,
	})

	filter := db.ChatbotFilter{}
	chatbots, _, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

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

func TestAdminListChatbots_NullFields(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	org := testdb.CreateOrganization(t, dbConn, testdb.OrganizationFixture{OwnerID: user.ID})
	ws := testdb.CreateWorkspace(t, dbConn, testdb.WorkspaceFixture{OrganizationID: org.Organization.ID})

	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{
		Name:           "Minimal Bot",
		UserID:         user.ID,
		WorkspaceID:    &ws.Workspace.ID,
		OrganizationID: &org.Organization.ID,
	})

	filter := db.ChatbotFilter{Name: ptr("Minimal")}
	chatbots, _, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	require.Len(t, chatbots, 1)

	bot := chatbots[0]
	assert.NotNil(t, bot.WorkspaceID, "workspace_id should be set")
	assert.NotNil(t, bot.OrganizationID, "organization_id should be set")
	assert.NotNil(t, bot.OrganizationName, "organization_name should be set")
}

func TestAdminListChatbots_SQLInjectionPrevention(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Normal Bot", UserID: user.ID})

	nameFilter := "' OR '1'='1"
	filter := db.ChatbotFilter{Name: &nameFilter}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 0, total, "SQL injection attempt should not match anything")
	assert.Len(t, chatbots, 0, "should return no results for SQL injection attempt")
}

func TestAdminListChatbots_SQLInjectionSpecialCharacters(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot With 'Quotes' & Symbols", UserID: user.ID})

	nameFilter := "'Quotes' & Symbols"
	filter := db.ChatbotFilter{Name: &nameFilter}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, total, "should find bot with special characters in name")
	assert.Len(t, chatbots, 1)
	assert.Equal(t, "Bot With 'Quotes' & Symbols", chatbots[0].Name)
}

func TestAdminListChatbots_EmptyResult(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	nameFilter := "NonExistentBot"
	filter := db.ChatbotFilter{Name: &nameFilter}

	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 0, total, "should return 0 total for no matches")
	assert.NotNil(t, chatbots, "should return empty slice, not nil")
	assert.Len(t, chatbots, 0, "should return empty slice")
}

func TestAdminListChatbots_DeletedExcluded(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	bot1 := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Active Bot", UserID: user.ID})
	bot2 := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Deleted Bot", UserID: user.ID})

	_, err := dbConn.ExecContext(ctx, `UPDATE chatbots SET deleted_at = NOW() WHERE id = $1`, bot2.Chatbot.ID)
	require.NoError(t, err, "failed to soft delete bot")

	filter := db.ChatbotFilter{}
	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, total, "should only count non-deleted chatbots")
	assert.Len(t, chatbots, 1)
	assert.Equal(t, bot1.Chatbot.ID, chatbots[0].ID, "should return only active bot")
}

func TestAdminListChatbots_NegativeLimit(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 2", UserID: user.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 3", UserID: user.ID})

	filter := db.ChatbotFilter{}
	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, -5, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, total, "should count all chatbots")
	assert.Len(t, chatbots, 3, "should default to limit of 20")
}

func TestAdminListChatbots_ZeroLimit(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID})
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 2", UserID: user.ID})

	filter := db.ChatbotFilter{}
	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 0, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "should count all chatbots")
	assert.Len(t, chatbots, 2, "should default to limit of 20")
}

func TestAdminListChatbots_NegativeOffset(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)
	_ = testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Bot 1", UserID: user.ID})

	filter := db.ChatbotFilter{}
	chatbots, total, err := db.AdminListChatbots(ctx, dbConn, filter, 10, -10)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, chatbots, 1, "should default to offset of 0")
}

func TestAdminListChatbots_SortByCreatedAtDesc(t *testing.T) {

	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	user := testdb.CreateUser(t, dbConn)

	bot1 := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "First Bot", UserID: user.ID})
	time.Sleep(5 * time.Millisecond)
	bot2 := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Second Bot", UserID: user.ID})
	time.Sleep(5 * time.Millisecond)
	bot3 := testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{Name: "Third Bot", UserID: user.ID})

	filter := db.ChatbotFilter{}
	chatbots, _, err := db.AdminListChatbots(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	require.Len(t, chatbots, 3)

	assert.Equal(t, bot3.Chatbot.ID, chatbots[0].ID, "first result should be newest (bot3)")
	assert.Equal(t, bot2.Chatbot.ID, chatbots[1].ID, "second result should be middle (bot2)")
	assert.Equal(t, bot1.Chatbot.ID, chatbots[2].ID, "third result should be oldest (bot1)")
}
