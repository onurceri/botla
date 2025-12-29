// Package repository defines interfaces for data access layer abstractions.
// These interfaces decouple business logic from database implementation,
// enabling easier testing and flexibility to swap storage backends.
package repository

import (
	"context"

	"github.com/onurceri/botla-co/internal/models"
)

// ChatbotRepository defines the interface for chatbot data access operations.
// Implementations must handle all CRUD operations for chatbots as well as
// workspace-scoped queries.
type ChatbotRepository interface {
	// GetByID retrieves a chatbot by its unique identifier.
	// Returns nil, nil if the chatbot is not found.
	GetByID(ctx context.Context, id string) (*models.Chatbot, error)

	// GetByUserID retrieves all non-deleted chatbots for a user.
	GetByUserID(ctx context.Context, userID string) ([]models.Chatbot, error)

	// GetByWorkspace retrieves all non-deleted chatbots for a workspace.
	GetByWorkspace(ctx context.Context, workspaceID string) ([]models.Chatbot, error)

	// Create persists a new chatbot and returns its generated ID.
	Create(ctx context.Context, bot *models.Chatbot) (string, error)

	// Update modifies an existing chatbot's fields.
	Update(ctx context.Context, bot *models.Chatbot) error

	// SoftDelete marks a chatbot as deleted and returns the IDs of associated sources.
	// This allows cleanup of related resources (e.g., vectors in Qdrant).
	SoftDelete(ctx context.Context, id, userID string) ([]string, error)

	// CountByUserID returns the number of active chatbots for a user.
	CountByUserID(ctx context.Context, userID string) (int, error)

	// CountByWorkspace returns the number of active chatbots for a workspace.
	CountByWorkspace(ctx context.Context, workspaceID string) (int, error)

	// UpdateSuggestedQuestions updates only the AI-generated suggestions.
	UpdateSuggestedQuestions(ctx context.Context, id string, suggestions []string) error
}

// ActionRepository defines the interface for chatbot action data access operations.
// Actions represent integrations (HTTP webhooks, Zapier, etc.) that chatbots can trigger.
type ActionRepository interface {
	// List returns all actions (enabled and disabled) for a chatbot, ordered by creation date descending.
	List(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error)

	// ListEnabled returns only enabled actions for a chatbot.
	ListEnabled(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error)

	// GetByID retrieves an action by its unique identifier.
	// Returns nil, nil if the action is not found.
	GetByID(ctx context.Context, id string) (*models.ChatbotAction, error)

	// GetByToolName finds an enabled action by its tool_name within a chatbot.
	// Returns nil, nil if no matching action is found.
	GetByToolName(ctx context.Context, chatbotID, toolName string) (*models.ChatbotAction, error)

	// Create persists a new action. The action's ID, Version, CreatedAt, and UpdatedAt
	// fields are populated after successful creation.
	Create(ctx context.Context, action *models.ChatbotAction) error

	// Update modifies an existing action with optimistic locking.
	// Returns ErrVersionConflict if the action was modified by another request.
	Update(ctx context.Context, action *models.ChatbotAction) error

	// Delete permanently removes an action by its ID.
	Delete(ctx context.Context, id string) error

	// GetLogs retrieves action execution logs for a chatbot with pagination.
	GetLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error)

	// CreateLog persists an action execution log entry.
	CreateLog(ctx context.Context, log *models.ActionExecutionLog) error
}

// SourceRepository defines the interface for data source access operations.
// Data sources represent the content (URLs, files) that feed a chatbot's knowledge base.
type SourceRepository interface {
	// GetByID retrieves a data source by its unique identifier.
	// Returns nil, nil if the source is not found.
	GetByID(ctx context.Context, id string) (*models.DataSource, error)

	// GetByChatbot retrieves all non-deleted data sources for a chatbot.
	GetByChatbot(ctx context.Context, chatbotID string) ([]models.DataSource, error)

	// GetURLSources retrieves all URL-type sources for a chatbot.
	GetURLSources(ctx context.Context, chatbotID string) ([]models.DataSource, error)

	// Create persists a new data source and returns its generated ID.
	Create(ctx context.Context, source *models.DataSource) (string, error)

	// SoftDelete marks a source as deleted by setting deleted_at timestamp.
	SoftDelete(ctx context.Context, id string) error

	// Delete permanently removes a source by its ID.
	Delete(ctx context.Context, id string) error

	// Exists checks if a source with the given URL already exists for a chatbot.
	Exists(ctx context.Context, chatbotID, url string) (bool, error)

	// ExistsByHash checks if a source with the same content hash exists for a chatbot.
	ExistsByHash(ctx context.Context, chatbotID, hash string) (bool, error)

	// GetByHash retrieves a source by its content hash within a chatbot.
	// Returns nil, nil if no matching source is found.
	GetByHash(ctx context.Context, chatbotID, hash string) (*models.DataSource, error)

	// CountByType counts non-deleted, non-failed sources of a specific type.
	CountByType(ctx context.Context, chatbotID, sourceType string) (int, error)
}

// AdminChatbot represents a chatbot for admin views with additional metadata.
// This type mirrors db.AdminChatbot to avoid direct db package dependency in handlers.
type AdminChatbot struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	OwnerID          string `json:"owner_id"`
	WorkspaceID      string `json:"workspace_id,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	OwnerEmail       string `json:"owner_email"`
	SourceCount      int    `json:"source_count"`
	MessageCount     int    `json:"message_count"`
	CustomBranding   []byte `json:"custom_branding"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// AdminChatbotFilter contains optional filters for listing chatbots in admin views.
type AdminChatbotFilter struct {
	Name           *string
	OrganizationID *string
	OwnerID        *string
}

// AdminChatbotRepository defines the interface for admin-specific chatbot data access operations.
// These methods are for platform administrators and have elevated access privileges.
type AdminChatbotRepository interface {
	// ListChatbots returns a paginated list of all chatbots with their metadata.
	// This is for admin views and includes details like owner email and message counts.
	ListChatbots(ctx context.Context, filter AdminChatbotFilter, limit, offset int) ([]AdminChatbot, int, error)

	// GetByID retrieves a single chatbot by ID with all admin-visible details.
	// Returns nil, sql.ErrNoRows if the chatbot is not found.
	GetByID(ctx context.Context, id string) (*AdminChatbot, error)

	// ResetSources resets all sources for a chatbot to pending status for reprocessing.
	// Returns the number of sources reset.
	ResetSources(ctx context.Context, chatbotID string) (int64, error)

	// GetSourceIDs returns all pending source IDs for a chatbot for queue processing.
	GetSourceIDs(ctx context.Context, chatbotID string) ([]string, error)

	// DeleteVectors resets chunk counts to 0 for all sources (for reindexing).
	DeleteVectors(ctx context.Context, chatbotID string) error
}

// ErrVersionConflict is returned when an optimistic lock fails due to concurrent modification.
// This typically occurs in Update operations when the entity has been modified
// by another request between read and write.
var ErrVersionConflict = repositoryError("version conflict: entity was modified by another request")

// repositoryError represents a repository-level error.
type repositoryError string

func (e repositoryError) Error() string {
	return string(e)
}
