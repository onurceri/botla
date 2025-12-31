// Package testdb provides test database utilities including fixture factories
// for creating test entities in isolation.
package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onurceri/botla-co/internal/auth"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/policy"
)

// =============================================================================
// User Fixtures
// =============================================================================

// UserFixture configures a test user's properties.
// All fields are optional; sensible defaults are applied via DefaultUserFixture.
type UserFixture struct {
	ID              string
	Email           string
	Password        string // Plain text, will be hashed
	FullName        string
	IsVerified      bool
	IsPlatformAdmin bool
	PlanCode        string // Plan code (e.g., policy.PlanFree, policy.PlanPro, policy.PlanUltra)
}

// DefaultUserFixture returns sensible defaults for a test user.
func DefaultUserFixture() UserFixture {
	return UserFixture{
		ID:              uuid.NewString(),
		Email:           fmt.Sprintf("test-%s@example.com", uuid.NewString()[:8]),
		Password:        "TestPassword123!",
		FullName:        "Test User",
		IsVerified:      true,
		IsPlatformAdmin: false,
		PlanCode:        policy.PlanFree.String(),
	}
}

// CreateUser creates a user in the test database.
// Accepts optional UserFixture to override defaults.
// Returns the created user model.
func CreateUser(t *testing.T, dbConn *sql.DB, fixture ...UserFixture) *models.User {
	t.Helper()

	f := DefaultUserFixture()
	if len(fixture) > 0 {
		f = mergeUserFixture(f, fixture[0])
	}

	ctx := context.Background()

	// Hash password
	hashedPassword, err := auth.HashPassword(f.Password)
	if err != nil {
		t.Fatalf("testdb.CreateUser: failed to hash password: %v", err)
	}

	// Get plan ID from code
	planID := getPlanID(t, dbConn, f.PlanCode)

	// Insert user
	var user models.User
	err = dbConn.QueryRowContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, plan_id, is_platform_admin, onboarding_completed, onboarding_step, onboarding_skipped)
		VALUES ($1, $2, $3, $4, $5, $6, true, 0, false)
		RETURNING id, email, full_name, plan_id, created_at, is_platform_admin
	`, f.ID, f.Email, hashedPassword, f.FullName, planID, f.IsPlatformAdmin).Scan(
		&user.ID, &user.Email, &user.FullName, &user.PlanID, &user.CreatedAt, &user.IsPlatformAdmin,
	)
	if err != nil {
		t.Fatalf("testdb.CreateUser: failed to create user: %v", err)
	}

	return &user
}

// mergeUserFixture merges override values into defaults.
// Zero values in override are ignored except for booleans.
func mergeUserFixture(defaults, override UserFixture) UserFixture {
	if override.ID != "" {
		defaults.ID = override.ID
	}
	if override.Email != "" {
		defaults.Email = override.Email
	}
	if override.Password != "" {
		defaults.Password = override.Password
	}
	if override.FullName != "" {
		defaults.FullName = override.FullName
	}
	if override.PlanCode != "" {
		defaults.PlanCode = override.PlanCode
	}
	// Boolean fields: always use override value (can explicitly set to false)
	defaults.IsVerified = override.IsVerified
	defaults.IsPlatformAdmin = override.IsPlatformAdmin
	return defaults
}

// =============================================================================
// Organization Fixtures
// =============================================================================

// OrganizationFixture configures a test organization's properties.
type OrganizationFixture struct {
	ID      string
	Name    string
	Slug    string
	OwnerID string // If empty, a new user is created
}

// DefaultOrganizationFixture returns sensible defaults for a test organization.
func DefaultOrganizationFixture() OrganizationFixture {
	suffix := uuid.NewString()[:8]
	return OrganizationFixture{
		ID:   uuid.NewString(),
		Name: "Test Organization " + suffix,
		Slug: "test-org-" + suffix,
	}
}

// OrganizationResult contains the created organization and its owner.
type OrganizationResult struct {
	Organization *models.Organization
	Owner        *models.User
}

// CreateOrganization creates an organization in the test database.
// If no OwnerID is provided in the fixture, a new user is created first.
// Returns the organization and its owner.
func CreateOrganization(t *testing.T, dbConn *sql.DB, fixture ...OrganizationFixture) *OrganizationResult {
	t.Helper()

	f := DefaultOrganizationFixture()
	if len(fixture) > 0 {
		f = mergeOrganizationFixture(f, fixture[0])
	}

	ctx := context.Background()

	// Create owner if not provided
	var owner *models.User
	if f.OwnerID == "" {
		owner = CreateUser(t, dbConn)
		f.OwnerID = owner.ID
	} else {
		// Load existing user
		var err error
		owner, err = db.GetUserByID(ctx, dbConn, f.OwnerID)
		if err != nil {
			t.Fatalf("testdb.CreateOrganization: failed to load owner %s: %v", f.OwnerID, err)
		}
		if owner == nil {
			t.Fatalf("testdb.CreateOrganization: owner not found: %s", f.OwnerID)
		}
	}

	// Create organization
	var org models.Organization
	err := dbConn.QueryRowContext(ctx, `
		INSERT INTO organizations (id, name, slug, owner_id, plan_id)
		VALUES ($1, $2, $3, $4, 'agency_starter')
		RETURNING id, name, slug, owner_id, plan_id, created_at, updated_at
	`, f.ID, f.Name, f.Slug, f.OwnerID).Scan(
		&org.ID, &org.Name, &org.Slug, &org.OwnerID, &org.PlanID, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("testdb.CreateOrganization: failed to create organization: %v", err)
	}

	// Add owner as member with 'owner' role
	_, err = dbConn.ExecContext(ctx, `
		INSERT INTO memberships (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, org.ID, f.OwnerID)
	if err != nil {
		t.Fatalf("testdb.CreateOrganization: failed to create membership: %v", err)
	}

	return &OrganizationResult{
		Organization: &org,
		Owner:        owner,
	}
}

// mergeOrganizationFixture merges override values into defaults.
func mergeOrganizationFixture(defaults, override OrganizationFixture) OrganizationFixture {
	if override.ID != "" {
		defaults.ID = override.ID
	}
	if override.Name != "" {
		defaults.Name = override.Name
	}
	if override.Slug != "" {
		defaults.Slug = override.Slug
	}
	if override.OwnerID != "" {
		defaults.OwnerID = override.OwnerID
	}
	return defaults
}

// =============================================================================
// Workspace Fixtures
// =============================================================================

// WorkspaceFixture configures a test workspace's properties.
type WorkspaceFixture struct {
	ID             string
	Name           string
	Slug           string
	OrganizationID string // If empty, a new organization is created
	ClientName     *string
}

// DefaultWorkspaceFixture returns sensible defaults for a test workspace.
func DefaultWorkspaceFixture() WorkspaceFixture {
	suffix := uuid.NewString()[:8]
	return WorkspaceFixture{
		ID:   uuid.NewString(),
		Name: "Test Workspace " + suffix,
		Slug: "test-ws-" + suffix,
	}
}

// WorkspaceResult contains the created workspace and its parent entities.
type WorkspaceResult struct {
	Workspace    *models.Workspace
	Organization *models.Organization
	Owner        *models.User
}

// CreateWorkspace creates a workspace in the test database.
// If no OrganizationID is provided, creates the full hierarchy.
// Returns the workspace and its parent organization and owner.
func CreateWorkspace(t *testing.T, dbConn *sql.DB, fixture ...WorkspaceFixture) *WorkspaceResult {
	t.Helper()

	f := DefaultWorkspaceFixture()
	if len(fixture) > 0 {
		f = mergeWorkspaceFixture(f, fixture[0])
	}

	ctx := context.Background()

	// Create org if not provided
	var org *models.Organization
	var owner *models.User
	if f.OrganizationID == "" {
		result := CreateOrganization(t, dbConn)
		org = result.Organization
		owner = result.Owner
		f.OrganizationID = org.ID
	}

	// Create workspace
	var ws models.Workspace
	err := dbConn.QueryRowContext(ctx, `
		INSERT INTO workspaces (id, organization_id, name, slug, client_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, organization_id, name, slug, client_name, created_at
	`, f.ID, f.OrganizationID, f.Name, f.Slug, f.ClientName).Scan(
		&ws.ID, &ws.OrganizationID, &ws.Name, &ws.Slug, &ws.ClientName, &ws.CreatedAt,
	)
	if err != nil {
		t.Fatalf("testdb.CreateWorkspace: failed to create workspace: %v", err)
	}

	return &WorkspaceResult{
		Workspace:    &ws,
		Organization: org,
		Owner:        owner,
	}
}

// mergeWorkspaceFixture merges override values into defaults.
func mergeWorkspaceFixture(defaults, override WorkspaceFixture) WorkspaceFixture {
	if override.ID != "" {
		defaults.ID = override.ID
	}
	if override.Name != "" {
		defaults.Name = override.Name
	}
	if override.Slug != "" {
		defaults.Slug = override.Slug
	}
	if override.OrganizationID != "" {
		defaults.OrganizationID = override.OrganizationID
	}
	if override.ClientName != nil {
		defaults.ClientName = override.ClientName
	}
	return defaults
}

// =============================================================================
// Chatbot Fixtures
// =============================================================================

// ChatbotFixture configures a test chatbot's properties.
type ChatbotFixture struct {
	ID                 string
	Name               string
	WorkspaceID        *string // If nil, a new workspace is created
	OrganizationID     *string
	UserID             string // If empty, uses the owner from workspace hierarchy
	Model              string
	LanguageCode       string
	WelcomeMessage     string
	Temperature        float32
	MaxTokens          int
	SuggestedQuestions []string
	SuggestionsEnabled bool
	SystemPrompt       string
}

// DefaultChatbotFixture returns sensible defaults for a test chatbot.
func DefaultChatbotFixture() ChatbotFixture {
	return ChatbotFixture{
		ID:             uuid.NewString(),
		Name:           "Test Bot " + uuid.NewString()[:8],
		Model:          policy.ModelGPT4oMini.String(),
		LanguageCode:   "en-US",
		WelcomeMessage: "Hello! How can I help you?",
		Temperature:    0.7,
		MaxTokens:      4096,
	}
}

// ChatbotResult contains the created chatbot and its parent entities.
type ChatbotResult struct {
	Chatbot      *models.Chatbot
	Workspace    *models.Workspace
	Organization *models.Organization
	User         *models.User
}

// CreateChatbot creates a chatbot with the full hierarchy (workspace, org, user).
// This is the most convenient fixture for tests that need a complete setup.
func CreateChatbot(t *testing.T, dbConn *sql.DB, fixture ...ChatbotFixture) *ChatbotResult {
	t.Helper()

	f := DefaultChatbotFixture()
	if len(fixture) > 0 {
		f = mergeChatbotFixture(f, fixture[0])
	}

	ctx := context.Background()

	// Create full hierarchy if not provided
	var workspace *models.Workspace
	var org *models.Organization
	var user *models.User

	if f.WorkspaceID == nil {
		wsResult := CreateWorkspace(t, dbConn)
		workspace = wsResult.Workspace
		org = wsResult.Organization
		user = wsResult.Owner
		f.WorkspaceID = &workspace.ID
		f.OrganizationID = &org.ID
		f.UserID = user.ID
	} else if f.UserID == "" {
		// If workspace provided but no user, create a user
		user = CreateUser(t, dbConn)
		f.UserID = user.ID
	}

	// Build the chatbot model for insertion
	bot := &models.Chatbot{
		ID:                   f.ID,
		Name:                 f.Name,
		SystemPrompt:         f.SystemPrompt,
		UserID:               f.UserID,
		WorkspaceID:          f.WorkspaceID,
		OrganizationID:       f.OrganizationID,
		Model:                f.Model,
		LanguageCode:         f.LanguageCode,
		WelcomeMessage:       f.WelcomeMessage,
		Temperature:          f.Temperature,
		MaxTokens:            f.MaxTokens,
		Position:             "bottom-right",
		ThemeColor:           "#ebb800",
		BotMessageColor:      "#f0f0f0",
		UserMessageColor:     "#ebb800",
		BotMessageTextColor:  "#000000",
		UserMessageTextColor: "#000000",
		ChatFontFamily:       "Inter",
		ChatHeaderColor:      "#ebb800",
		ChatHeaderTextColor:  "#000000",
		ChatBackgroundColor:  "#ffffff",
		BubbleRadius:         "22px",
		InputBackgroundColor: "#ededed",
		InputTextColor:       "#000000",
		SendButtonColor:      "#ebb800",
		DiscoveryMode:        "auto",
		RefreshPolicy:        "manual",
		ConfidenceThreshold:  0.5,
		HandoffType:          "email",
		SuggestedQuestions:   f.SuggestedQuestions,
		SuggestionsEnabled:   f.SuggestionsEnabled,
	}

	// Use the db package's CreateChatbot function
	id, err := db.CreateChatbot(ctx, dbConn, bot)
	if err != nil {
		t.Fatalf("testdb.CreateChatbot: failed to create chatbot: %v", err)
	}

	// Reload the created chatbot to get all fields
	chatbot, err := db.GetChatbotByID(ctx, dbConn, id)
	if err != nil {
		t.Fatalf("testdb.CreateChatbot: failed to reload chatbot: %v", err)
	}

	return &ChatbotResult{
		Chatbot:      chatbot,
		Workspace:    workspace,
		Organization: org,
		User:         user,
	}
}

// mergeChatbotFixture merges override values into defaults.
func mergeChatbotFixture(defaults, override ChatbotFixture) ChatbotFixture {
	if override.ID != "" {
		defaults.ID = override.ID
	}
	if override.Name != "" {
		defaults.Name = override.Name
	}
	if override.WorkspaceID != nil {
		defaults.WorkspaceID = override.WorkspaceID
	}
	if override.OrganizationID != nil {
		defaults.OrganizationID = override.OrganizationID
	}
	if override.UserID != "" {
		defaults.UserID = override.UserID
	}
	if override.Model != "" {
		defaults.Model = override.Model
	}
	if override.LanguageCode != "" {
		defaults.LanguageCode = override.LanguageCode
	}
	if override.WelcomeMessage != "" {
		defaults.WelcomeMessage = override.WelcomeMessage
	}
	if override.Temperature != 0 {
		defaults.Temperature = override.Temperature
	}
	if override.MaxTokens != 0 {
		defaults.MaxTokens = override.MaxTokens
	}
	if len(override.SuggestedQuestions) > 0 {
		defaults.SuggestedQuestions = override.SuggestedQuestions
	}
	if override.SuggestionsEnabled {
		defaults.SuggestionsEnabled = override.SuggestionsEnabled
	}
	if override.SystemPrompt != "" {
		defaults.SystemPrompt = override.SystemPrompt
	}
	return defaults
}

// =============================================================================
// Source Fixtures
// =============================================================================

// SourceFixture configures a test data source's properties.
type SourceFixture struct {
	ID         string
	ChatbotID  string // If empty, a new chatbot is created
	SourceType string // "text", "url", "file"
	SourceURL  *string
	FilePath   *string
	Status     string
	ChunkCount int
}

// DefaultSourceFixture returns sensible defaults for a test source.
func DefaultSourceFixture() SourceFixture {
	return SourceFixture{
		ID:         uuid.NewString(),
		SourceType: "text",
		Status:     "completed",
		ChunkCount: 1,
	}
}

// SourceResult contains the created source and its parent entities.
type SourceResult struct {
	Source       *models.DataSource
	Chatbot      *models.Chatbot
	Workspace    *models.Workspace
	Organization *models.Organization
	User         *models.User
}

// CreateSource creates a data source with the full hierarchy.
func CreateSource(t *testing.T, dbConn *sql.DB, fixture ...SourceFixture) *SourceResult {
	t.Helper()

	f := DefaultSourceFixture()
	if len(fixture) > 0 {
		f = mergeSourceFixture(f, fixture[0])
	}

	ctx := context.Background()

	// Create chatbot if not provided
	var chatbot *models.Chatbot
	var workspace *models.Workspace
	var org *models.Organization
	var user *models.User

	if f.ChatbotID == "" {
		cbResult := CreateChatbot(t, dbConn)
		chatbot = cbResult.Chatbot
		workspace = cbResult.Workspace
		org = cbResult.Organization
		user = cbResult.User
		f.ChatbotID = chatbot.ID
	}

	// Create source
	var source models.DataSource
	now := time.Now()
	err := dbConn.QueryRowContext(ctx, `
		INSERT INTO data_sources (id, chatbot_id, source_type, source_url, file_path, status, chunk_count, processed_at, size_bytes, is_discovered)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, false)
		RETURNING id, chatbot_id, source_type, source_url, file_path, status, chunk_count, processed_at, created_at
	`, f.ID, f.ChatbotID, f.SourceType, f.SourceURL, f.FilePath, f.Status, f.ChunkCount, now).Scan(
		&source.ID, &source.ChatbotID, &source.SourceType, &source.SourceURL, &source.FilePath,
		&source.Status, &source.ChunkCount, &source.ProcessedAt, &source.CreatedAt,
	)
	if err != nil {
		t.Fatalf("testdb.CreateSource: failed to create source: %v", err)
	}

	return &SourceResult{
		Source:       &source,
		Chatbot:      chatbot,
		Workspace:    workspace,
		Organization: org,
		User:         user,
	}
}

// mergeSourceFixture merges override values into defaults.
func mergeSourceFixture(defaults, override SourceFixture) SourceFixture {
	if override.ID != "" {
		defaults.ID = override.ID
	}
	if override.ChatbotID != "" {
		defaults.ChatbotID = override.ChatbotID
	}
	if override.SourceType != "" {
		defaults.SourceType = override.SourceType
	}
	if override.SourceURL != nil {
		defaults.SourceURL = override.SourceURL
	}
	if override.FilePath != nil {
		defaults.FilePath = override.FilePath
	}
	if override.Status != "" {
		defaults.Status = override.Status
	}
	if override.ChunkCount != 0 {
		defaults.ChunkCount = override.ChunkCount
	}
	return defaults
}

// =============================================================================
// Helper Functions
// =============================================================================

// getPlanID retrieves the plan ID for the given plan code.
func getPlanID(t *testing.T, dbConn *sql.DB, planCode string) string {
	t.Helper()

	var planID string
	err := dbConn.QueryRowContext(context.Background(),
		`SELECT id FROM plans WHERE code = $1`, planCode).Scan(&planID)
	if err != nil {
		// Fall back to first available plan
		err = dbConn.QueryRowContext(context.Background(),
			`SELECT id FROM plans LIMIT 1`).Scan(&planID)
		if err != nil {
			t.Fatalf("testdb.getPlanID: no plans found in database (migrations might not have run): %v", err)
		}
	}
	return planID
}

// StringPtr is a helper to create a pointer to a string value.
func StringPtr(s string) *string {
	return &s
}
