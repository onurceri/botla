package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserDataForExport_Content(t *testing.T) {
	pool := testdb.OpenTestDB(t)
	ctx := context.Background()

	// 1. Create a user
	userID := createExportTestUser(t, pool)

	// 2. Create an organization and membership
	var orgID string
	err := pool.QueryRow(`
		INSERT INTO organizations (name, slug, owner_id)
		VALUES ('Test Org', 'test-org-' || $1, $1::uuid)
		RETURNING id
	`, userID).Scan(&orgID)
	require.NoError(t, err)

	_, err = pool.Exec(`
		INSERT INTO memberships (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, orgID, userID)
	require.NoError(t, err)

	// 3. Create a chatbot
	chatbot := &models.Chatbot{
		UserID:         userID,
		OrganizationID: &orgID,
		Name:           "Export Test Bot",
		SystemPrompt:   "You are a test bot",
		Model:          "gpt-3.5-turbo",
	}
	chatbotID, err := CreateChatbot(ctx, pool, chatbot)
	require.NoError(t, err)

	// 4. Create a conversation
	conv, err := GetOrCreateConversationBySessionID(ctx, pool, chatbotID, "session-123")
	require.NoError(t, err)

	// 5. Create a consent
	err = UpsertConsent(ctx, pool, userID, "marketing", true, "127.0.0.1", "test-agent")
	require.NoError(t, err)

	// Now run the export logic
	export, err := GetUserDataForExport(ctx, pool, userID)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, userID, export.User.ID)

	assert.Len(t, export.Organizations, 1)
	assert.Equal(t, "Test Org", export.Organizations[0].Name)

	assert.Len(t, export.Chatbots, 1)
	assert.Equal(t, "Export Test Bot", export.Chatbots[0].Name)

	assert.Len(t, export.Conversations, 1)
	assert.Equal(t, conv.ID, export.Conversations[0].ID)

	assert.Len(t, export.Consents, 1)
	assert.Equal(t, "marketing", export.Consents[0].ConsentType)
}

func createExportTestUser(t *testing.T, db *sql.DB) string {
	t.Helper()
	email := fmt.Sprintf("export_test_%d@example.com", time.Now().UnixNano())
	var id string
	var freePlanID string
	if err := db.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&freePlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", freePlanID).Scan(&id); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}
