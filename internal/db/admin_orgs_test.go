package db_test

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminListOrganizations(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	// 1. Setup test data
	// Create an organization (this also creates an owner user)
	res := testdb.CreateOrganization(t, dbConn, testdb.OrganizationFixture{
		Name: "Test Org",
	})
	orgID := res.Organization.ID
	userID := res.Owner.ID

	// Add another member
	user2 := testdb.CreateUser(t, dbConn)
	_, err := dbConn.Exec(`INSERT INTO memberships (organization_id, user_id, role) VALUES ($1, $2, $3)`, orgID, user2.ID, "member")
	require.NoError(t, err)

	// Create a workspace
	ws := testdb.CreateWorkspace(t, dbConn, testdb.WorkspaceFixture{
		OrganizationID: orgID,
	})

	// Add a chatbot
	testdb.CreateChatbot(t, dbConn, testdb.ChatbotFixture{
		Name:           "Bot 1",
		WorkspaceID:    &ws.Workspace.ID,
		OrganizationID: &orgID,
		UserID:         userID,
	})

	// 2. Test AdminListOrganizations
	filter := db.OrganizationFilter{}
	orgs, total, err := db.AdminListOrganizations(ctx, dbConn, filter, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)

	var found bool
	for _, o := range orgs {
		if o.ID == orgID {
			found = true
			assert.Equal(t, 2, o.UserCount, "Should have 2 users (owner + 1 member)")
			assert.Equal(t, 1, o.ChatbotCount, "Should have 1 chatbot")
		}
	}
	assert.True(t, found, "Test organization should be found")

	// 3. Test GetOrganizationByID
	org, err := db.GetOrganizationByID(ctx, dbConn, orgID)
	require.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, 2, org.UserCount)
	assert.Equal(t, 1, org.ChatbotCount)
}
