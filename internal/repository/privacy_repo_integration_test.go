package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

func TestPostgresPrivacyRepo_GetUserDataForExport_Integration(t *testing.T) {
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("Skipping integration test; set TEST_INTEGRATION=1 to run")
	}

	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresPrivacyRepo(db)
	ctx := context.Background()

	// Create test user
	userID := "test-user-" + time.Now().Format("20060102150405")
	email := userID + "@example.com"

	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, full_name, created_at)
		VALUES ($1, $2, 'Test User', NOW())
	`, userID, email)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create organization and membership
	orgID := "test-org-" + time.Now().Format("20060102150405")
	_, err = db.ExecContext(ctx, `
		INSERT INTO organizations (id, name, created_at)
		VALUES ($1, 'Test Organization', NOW())
	`, orgID)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM organizations WHERE id = $1", orgID)

	_, err = db.ExecContext(ctx, `
		INSERT INTO memberships (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, orgID, userID)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM memberships WHERE user_id = $1", userID)

	// Create chatbot
	botID := "test-bot-" + time.Now().Format("20060102150405")
	_, err = db.ExecContext(ctx, `
		INSERT INTO chatbots (id, user_id, name, created_at)
		VALUES ($1, $2, 'Test Bot', NOW())
	`, botID, userID)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM chatbots WHERE id = $1", botID)

	// Create conversation
	convID := "test-conv-" + time.Now().Format("20060102150405")
	_, err = db.ExecContext(ctx, `
		INSERT INTO conversations (id, chatbot_id, session_id, created_at)
		VALUES ($1, $2, 'session-123', NOW())
	`, convID, botID)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM conversations WHERE id = $1", convID)

	// Create messages
	msgID1 := "test-msg1-" + time.Now().Format("20060102150405")
	msgID2 := "test-msg2-" + time.Now().Format("20060102150405")
	_, err = db.ExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, role, content, created_at)
		VALUES ($1, $2, 'user', 'Hello', NOW())
	`, msgID1, convID)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM messages WHERE id = $1", msgID1)

	_, err = db.ExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, role, content, created_at)
		VALUES ($1, $2, 'assistant', 'Hi there!', NOW())
	`, msgID2, convID)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM messages WHERE id = $1", msgID2)

	// Test GetUserDataForExport
	export, err := repo.GetUserDataForExport(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, export)

	// Verify user data
	assert.Equal(t, userID, export.User.ID)
	assert.Equal(t, email, export.User.Email)

	// Verify organizations
	assert.Len(t, export.Organizations, 1)
	assert.Equal(t, "Test Organization", export.Organizations[0].Name)

	// Verify chatbots
	assert.Len(t, export.Chatbots, 1)
	assert.Equal(t, "Test Bot", export.Chatbots[0].Name)

	// Verify conversations
	assert.Len(t, export.Conversations, 1)
	assert.Equal(t, botID, export.Conversations[0].ChatbotID)

	// Verify messages
	assert.Len(t, export.Messages, 2)

	// Verify exported_at is set
	assert.False(t, export.ExportedAt.IsZero())
}

func TestPostgresPrivacyRepo_GetUserDataForExport_EmptyUser_Integration(t *testing.T) {
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("Skipping integration test; set TEST_INTEGRATION=1 to run")
	}

	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresPrivacyRepo(db)
	ctx := context.Background()

	// Create test user with no related data
	userID := "test-empty-user-" + time.Now().Format("20060102150405")
	email := userID + "@example.com"

	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, full_name, created_at)
		VALUES ($1, $2, 'Empty User', NOW())
	`, userID, email)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Test GetUserDataForExport
	export, err := repo.GetUserDataForExport(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, export)

	// Verify user data
	assert.Equal(t, userID, export.User.ID)
	assert.Equal(t, email, export.User.Email)

	// Verify empty collections
	assert.Empty(t, export.Organizations)
	assert.Empty(t, export.Chatbots)
	assert.Empty(t, export.Conversations)
	assert.Empty(t, export.Messages)
	assert.Empty(t, export.ActionLogs)
	assert.Empty(t, export.Consents)
}

func TestPostgresPrivacyRepo_CompletePrivacyExportRequest_Integration(t *testing.T) {
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("Skipping integration test; set TEST_INTEGRATION=1 to run")
	}

	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresPrivacyRepo(db)
	ctx := context.Background()

	// Create admin user
	adminID := "test-admin-" + time.Now().Format("20060102150405")
	adminEmail := adminID + "@example.com"
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, full_name, is_platform_admin, created_at)
		VALUES ($1, $2, 'Admin User', true, NOW())
	`, adminID, adminEmail)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", adminID)

	// Create privacy request
	requestID := "test-req-" + time.Now().Format("20060102150405")
	_, err = db.ExecContext(ctx, `
		INSERT INTO privacy_requests (id, user_email, request_type, status, created_at)
		VALUES ($1, 'user@example.com', 'export', 'processing', NOW())
	`, requestID)
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM privacy_requests WHERE id = $1", requestID)

	// Test CompletePrivacyExportRequest
	exportURL := "exports/test/export.json"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	err = repo.CompletePrivacyExportRequest(ctx, requestID, adminID, exportURL, expiresAt)
	require.NoError(t, err)

	// Verify the update
	var status, savedExportURL string
	var processedBy *string
	var savedExpiresAt time.Time
	err = db.QueryRowContext(ctx, `
		SELECT status, export_url, processed_by, export_expires_at
		FROM privacy_requests WHERE id = $1
	`, requestID).Scan(&status, &savedExportURL, &processedBy, &savedExpiresAt)
	require.NoError(t, err)

	assert.Equal(t, "completed", status)
	assert.Equal(t, exportURL, savedExportURL)
	assert.NotNil(t, processedBy)
	assert.Equal(t, adminID, *processedBy)
	assert.WithinDuration(t, expiresAt, savedExpiresAt, time.Second)
}
