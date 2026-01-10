// Package repository defines interfaces for data access layer abstractions.
// These interfaces decouple business logic from database implementation,
// enabling easier testing and flexibility to swap storage backends.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/onurceri/botla-app/internal/models"
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

	// GetDueForRefresh returns chatbots with auto refresh enabled that are due for refresh.
	GetDueForRefresh(ctx context.Context, now time.Time) ([]models.Chatbot, error)

	// UpdateRefreshTimes updates the next_refresh_at and last_refresh_at for a chatbot.
	UpdateRefreshTimes(ctx context.Context, botID string, nextRefresh, lastRefresh time.Time) error
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

	// UpdateForRefresh sets status to pending and clears error_message for a source refresh.
	UpdateForRefresh(ctx context.Context, id string) error

	// UpdateSourceHash updates the content hash for a source.
	UpdateSourceHash(ctx context.Context, id string, hash string) error

	// UpdateSourceProcessing updates processing status, error, chunk count, and processed_at.
	UpdateSourceProcessing(ctx context.Context, id string, status string, errorMessage *string, chunkCount int, processedAt *time.Time) error

	// UpdateSourceCapability updates the capability summary for a source.
	UpdateSourceCapability(ctx context.Context, id string, summary string) error

	// UpdateSourceSuggestions updates the suggested questions for a source.
	UpdateSourceSuggestions(ctx context.Context, id string, suggestions []string) error

	// GetLastDeletedAtForURL returns the most recent deleted_at timestamp for a given URL in a chatbot.
	// Returns sql.NullTime{} if no deleted source is found.
	GetLastDeletedAtForURL(ctx context.Context, chatbotID, url string) (time.Time, bool, error)
}

// AdminChatbot represents a chatbot for admin views with additional metadata.
// This type mirrors db.AdminChatbot to avoid direct db package dependency in handlers.
type AdminChatbot struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	OwnerID          string  `json:"owner_id"`
	WorkspaceID      *string `json:"workspace_id,omitempty"`
	OrganizationID   *string `json:"organization_id,omitempty"`
	OrganizationName *string `json:"organization_name,omitempty"`
	OwnerEmail       string  `json:"owner_email"`
	SourceCount      int     `json:"source_count"`
	MessageCount     int     `json:"message_count"`
	CustomBranding   []byte  `json:"custom_branding"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
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
var ErrVersionConflict = errors.New("version conflict: entity was modified by another request")

// PlanRepository defines the interface for plan data access operations.
// Plans define feature limits and pricing tiers for users.
// Implementations should support Redis caching for GetByUserID operations.
type PlanRepository interface {
	// GetByUserID retrieves the active plan for a user.
	// Returns nil if user has no plan or is not found.
	// Results are cached in Redis for performance when a Redis client is available.
	GetByUserID(ctx context.Context, userID string) (*models.Plan, error)

	// GetByCode retrieves a plan by its code (e.g., "free", "pro", "ultra").
	// Returns nil, nil if no active plan with that code exists.
	GetByCode(ctx context.Context, code string) (*models.Plan, error)

	// GetAll retrieves all active plans ordered by price ascending.
	GetAll(ctx context.Context) ([]models.Plan, error)

	// GetByID retrieves a plan by its unique identifier.
	GetByID(ctx context.Context, id string) (*models.Plan, error)

	// GetPlanWithLimits retrieves a plan by user ID with all limits populated.
	GetPlanWithLimits(ctx context.Context, userID string) (*models.Plan, error)

	// GetAllPlansWithLimits retrieves all active plans with their limits.
	GetAllPlansWithLimits(ctx context.Context) ([]models.Plan, error)

	// InvalidateCache removes the cached plan for a user.
	// Call this when a user's plan changes (upgrade/downgrade).
	InvalidateCache(ctx context.Context, userID string) error
}

// ConversationRepository defines the interface for conversation data access operations.
// Conversations contain chat sessions between users and chatbots.
type ConversationRepository interface {
	// GetOrCreateBySessionID finds an existing conversation or creates a new one.
	// Uses session_id as the unique identifier within a chatbot.
	GetOrCreateBySessionID(ctx context.Context, chatbotID, sessionID string) (*models.Conversation, error)

	// GetByID retrieves a conversation by its unique identifier.
	GetByID(ctx context.Context, id string) (*models.Conversation, error)

	// CreateMessage persists a new message in a conversation.
	// Returns the generated message ID.
	CreateMessage(ctx context.Context, msg *models.Message) (string, error)

	// GetMessages retrieves messages for a conversation with pagination.
	// Messages are ordered by created_at ascending (chronological order).
	GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]models.Message, error)

	// IncrementMessageCount atomically increments the message count for a conversation.
	IncrementMessageCount(ctx context.Context, conversationID string) error

	// ListRecentMessages retrieves recent messages for a conversation.
	ListRecentMessages(ctx context.Context, conversationID string, limit int) ([]models.Message, error)

	// SaveMessageSources persists source usage for a message.
	SaveMessageSources(ctx context.Context, messageID string, sources []models.ChunkMetadata) error
}

// ChunkMetadata represents metadata about a source chunk used in RAG.
type ChunkMetadata struct {
	ChunkIndex  int    `json:"chunk_index"`
	SourceType  string `json:"source_type"`
	SourceID    string `json:"source_id"`
	ContentHash string `json:"content_hash"`
}

// AnalyticsRepository defines the interface for analytics data access operations.
type AnalyticsRepository interface {
	// GetOverview returns aggregated stats for a chatbot for the last 30 days.
	GetOverview(ctx context.Context, chatbotID string) (*models.AnalyticsOverview, error)

	// GetTrends returns daily stats for a chatbot for the last N days.
	GetTrends(ctx context.Context, chatbotID string, days int) ([]models.DailyAnalytics, error)

	// IncrementAnalytics updates analytics counters for a chatbot.
	IncrementAnalytics(ctx context.Context, chatbotID string, isNewConversation bool, tokens int, isHandoff bool, responseTimeMs int) error

	// TrackUnansweredQuery records a query that had low confidence.
	TrackUnansweredQuery(ctx context.Context, chatbotID, queryText string) error

	// GetMonthlyTokenUsage returns the total tokens used by all chatbots of a user in the current month.
	GetMonthlyTokenUsage(ctx context.Context, userID string) (int, error)

	// GetAutoRefreshCountForMonth returns the auto_refresh_count for a user in a given month.
	GetAutoRefreshCountForMonth(ctx context.Context, userID string, month time.Time) (int, error)

	// IncrementAutoRefreshCount increments the auto_refresh_count for a user in a given month.
	IncrementAutoRefreshCount(ctx context.Context, userID string, month time.Time, delta int) error

	// GetGlobalAnalytics returns aggregated analytics for a scope (User, Workspace, or Org).
	GetGlobalAnalytics(ctx context.Context, userID string, orgID, wsID *string) ([]AnalyticsPoint, error)

	// GetSourceUsageStats returns source usage statistics for a chatbot.
	GetSourceUsageStats(ctx context.Context, chatbotID string, days int) ([]SourceUsageStat, error)

	// IncrementFeedback updates positive or negative feedback counters.
	IncrementFeedback(ctx context.Context, chatbotID string, oldThumbsUp *bool, newThumbsUp bool) error

	// UpdateMessageFeedback updates feedback for a message and returns affected chatbot ID.
	UpdateMessageFeedback(ctx context.Context, messageID string, thumbsUp bool) (string, bool, error)
}

// PrivacyRepository defines the interface for privacy data access operations.
type PrivacyRepository interface {
	// CreateDataExport creates a new data export record.
	CreateDataExport(ctx context.Context, exp DataExport) (*DataExport, error)

	// UpdateDataExport updates an existing data export record.
	UpdateDataExport(ctx context.Context, exp DataExport) error

	// GetDataExport retrieves a data export by ID.
	GetDataExport(ctx context.Context, id string) (*DataExport, error)

	// GetUserDataForExport retrieves all user data for GDPR export.
	GetUserDataForExport(ctx context.Context, userID string) (*UserDataExport, error)

	// CreatePrivacyRequest creates a new privacy request.
	CreatePrivacyRequest(ctx context.Context, req PrivacyRequest) (*PrivacyRequest, error)

	// GetPrivacyRequest retrieves a privacy request by ID.
	GetPrivacyRequest(ctx context.Context, requestID string) (*PrivacyRequest, error)

	// ListPrivacyRequests retrieves privacy requests with optional status filter and pagination.
	ListPrivacyRequests(ctx context.Context, status string, limit, offset int) ([]PrivacyRequest, int, error)

	// UpdatePrivacyRequestStatus updates the status of a privacy request.
	UpdatePrivacyRequestStatus(ctx context.Context, requestID, status, adminID string, denialReason *string) error

	// CompletePrivacyExportRequest marks a privacy export request as completed.
	CompletePrivacyExportRequest(ctx context.Context, requestID, adminID, exportURL string, expiresAt time.Time) error

	// AnonymizeUserData anonymizes a user's personal data and deletes their content.
	AnonymizeUserData(ctx context.Context, userID string) error

	// GetUserFilesForDeletion returns file paths that should be deleted from storage.
	GetUserFilesForDeletion(ctx context.Context, userID string) ([]string, error)

	// GetUserConsents retrieves all consent records for a user.
	GetUserConsents(ctx context.Context, userID string) ([]UserConsent, error)

	// UpsertConsent creates or updates a consent record.
	UpsertConsent(ctx context.Context, userID string, consentType string, granted bool, ipAddress, userAgent string) error

	// ListPrivacyRequestsByUserID retrieves privacy requests for a specific user with pagination and optional request type filter.
	ListPrivacyRequestsByUserID(ctx context.Context, userID, requestType string, limit, offset int) ([]PrivacyRequest, int, error)

	// HasActivePrivacyRequest checks if user has a pending or processing request of the given type.
	HasActivePrivacyRequest(ctx context.Context, userID, requestType string) (bool, error)

	// DeletePrivacyRequest deletes a privacy request by ID.
	DeletePrivacyRequest(ctx context.Context, requestID string) error

	// GetLastCompletedRequestDate returns the completion date of the last completed request of the given type for a user.
	GetLastCompletedRequestDate(ctx context.Context, userID, requestType string) (*time.Time, error)
}

// HandoffRepository defines the interface for human handoff data access operations.
type HandoffRepository interface {
	// HasActiveHandoffRequest checks if there is any pending or assigned handoff request for a conversation.
	HasActiveHandoffRequest(ctx context.Context, conversationID string) (bool, error)

	// CreateHandoffRequest creates a new handoff request and returns the generated ID.
	CreateHandoffRequest(ctx context.Context, req *models.HandoffRequest) (string, error)

	// GetHandoffRequestsByBotID returns all handoff requests for a chatbot.
	GetHandoffRequestsByBotID(ctx context.Context, chatbotID string) ([]*models.HandoffRequest, error)

	// GetHandoffRequestByID returns a single handoff request by ID.
	GetHandoffRequestByID(ctx context.Context, id string) (*models.HandoffRequest, error)

	// UpdateHandoffRequestStatus updates the status of a handoff request.
	UpdateHandoffRequestStatus(ctx context.Context, id, status string, assignedTo *string) error

	// ListHandoffMessages retrieves conversation messages for a handoff request.
	ListHandoffMessages(ctx context.Context, conversationID string, limit int) ([]models.Message, error)
}

// UserFilter contains optional filters for listing users in admin views.
type UserFilter struct {
	Email           *string
	IsPlatformAdmin *bool
	PlanID          *string
}

// UserRepository defines the interface for user data access operations.
type UserRepository interface {
	// GetByID retrieves a user by their unique identifier.
	// Returns nil, nil if the user is not found.
	GetByID(ctx context.Context, id string) (*models.User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// AdminListUsers returns a paginated list of users for admin views.
	AdminListUsers(ctx context.Context, filter UserFilter, limit, offset int) ([]*models.User, int, error)

	// AdminUpdateUser updates a user's fields for admin operations.
	AdminUpdateUser(ctx context.Context, id string, updates map[string]any) error

	// UpdateOnboardingState updates the user's onboarding progress.
	UpdateOnboardingState(ctx context.Context, userID string, step int, data *models.OnboardingData) error

	// SkipOnboarding marks the user's onboarding as skipped.
	SkipOnboarding(ctx context.Context, userID string) error

	// CompleteOnboarding marks the user's onboarding as completed.
	CompleteOnboarding(ctx context.Context, userID, botID string) error

	// GetTotalUsers returns the total count of non-deleted users.
	GetTotalUsers(ctx context.Context) (int, error)

	// GetTotalMessages returns the total count of messages.
	GetTotalMessages(ctx context.Context) (int, error)
}

// UsageRepository defines the interface for usage/ingestion data access operations.
type UsageRepository interface {
	CountChatbotsByUserID(ctx context.Context, userID string) (int, error)
	CountChatbotsByWorkspace(ctx context.Context, workspaceID string) (int, error)
	GetFileCountByUserID(ctx context.Context, userID string) (int, error)
	GetURLCountByUserID(ctx context.Context, userID string) (int, error)
	GetStorageUsedMBByUserID(ctx context.Context, userID string) (int, error)
	GetMaxFileCountInAnyBot(ctx context.Context, userID string) (int, error)
	GetMaxURLCountInAnyBot(ctx context.Context, userID string) (int, error)
	GetMonthlyTokenUsage(ctx context.Context, userID string) (int, error)
	GetMonthlyIngestionUsage(ctx context.Context, userID string, month time.Time) (int, int, error)
	GetMonthlyRefreshCount(ctx context.Context, userID string, month time.Time) (int, error)
	IncrementRefreshCount(ctx context.Context, userID string, month time.Time) error
	IncrementSuccessfulIngestion(ctx context.Context, userID string, at time.Time, delta int) error
	AddEmbeddingTokens(ctx context.Context, userID string, at time.Time, tokens int) error
	GetAutoRefreshCountForMonth(ctx context.Context, userID string, month time.Time) (int, error)
	IncrementAutoRefreshCount(ctx context.Context, userID string, month time.Time, delta int) error

	// ReserveChatTokens atomically reserves tokens for a chat request.
	// Returns ErrTokenQuotaExceeded if the reservation would exceed the limit.
	ReserveChatTokens(ctx context.Context, userID string, estimatedTokens int, maxMonthlyTokens int) error

	// AdjustChatTokens adjusts the token count after a chat request completes.
	AdjustChatTokens(ctx context.Context, userID string, deltaTokens int) error

	// GetMonthlyChatTokens returns the current monthly chat token usage.
	GetMonthlyChatTokens(ctx context.Context, userID string) (int, error)

	// IncrementChatTokens adds to the chat_tokens counter (no limit check).
	IncrementChatTokens(ctx context.Context, userID string, tokens int) error
}

// QueueRepository defines the interface for queue administration operations.
type QueueRepository interface {
	GetQueueStats(ctx context.Context) ([]QueueStats, error)
	GetStuckJobs(ctx context.Context, threshold time.Duration) ([]StuckJob, error)
}

// TrainingJobRepository defines the interface for training job data access operations.
type TrainingJobRepository interface {
	GetByID(ctx context.Context, id string) (*models.TrainingJob, error)
	GetBySourceID(ctx context.Context, sourceID string) (*models.TrainingJob, error)
	GetByChatbotID(ctx context.Context, chatbotID string, limit int) ([]*models.TrainingJob, error)
	Create(ctx context.Context, sourceID, chatbotID string) (*models.TrainingJob, error)
	UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, step *models.TrainingStep) error
	ResetForRetry(ctx context.Context, id string) error

	// IncrementRetryCount increments the retry count for a job.
	IncrementRetryCount(ctx context.Context, id string) (int, error)

	// GetPendingJobs retrieves jobs in pending status for recovery.
	GetPendingJobs(ctx context.Context, limit int) ([]*models.TrainingJob, error)

	// MarkStepCompleted marks a step as completed in job metadata.
	MarkStepCompleted(ctx context.Context, jobID string, step models.TrainingStep, outputHash string) error

	// GetLastCompletedStep returns the last completed step for resuming.
	GetLastCompletedStep(ctx context.Context, jobID string) (*models.TrainingStep, error)

	// Fail marks a job as failed with error details.
	Fail(ctx context.Context, id string, step models.TrainingStep, errCode, errMsg string) error

	// Complete marks a job as completed.
	Complete(ctx context.Context, id string) error

	// Cancel marks a job as cancelled.
	Cancel(ctx context.Context, id string) error

	// GetRetryableJobs retrieves failed jobs that can be retried.
	GetRetryableJobs(ctx context.Context, maxRetries, limit int) ([]*models.TrainingJob, error)

	// GetRunningJobs retrieves jobs currently running.
	GetRunningJobs(ctx context.Context, limit int) ([]*models.TrainingJob, error)
}

// SuggestionJobRepository defines the interface for suggestion job data access operations.
type SuggestionJobRepository interface {
	Create(ctx context.Context, chatbotID string) (*models.SuggestionJob, error)
	GetByID(ctx context.Context, id string) (*models.SuggestionJob, error)
	GetLatestForChatbot(ctx context.Context, chatbotID string) (*models.SuggestionJob, error)
	UpdateStatus(ctx context.Context, id string, status models.SuggestionJobStatus) error
	Complete(ctx context.Context, id string, suggestions []string) error
	Fail(ctx context.Context, id string, errMsg string) error
}

// PendingURLRepository defines the interface for pending discovered URL operations.
type PendingURLRepository interface {
	// InsertPendingURL adds a URL to the pending list for approval.
	InsertPendingURL(ctx context.Context, chatbotID string, sourceID *string, url string) error

	// ListPendingURLs returns pending URLs for a chatbot with pagination.
	ListPendingURLs(ctx context.Context, chatbotID string, limit, offset int) ([]models.PendingURL, error)

	// CountPendingURLs returns the total count of pending URLs for a chatbot.
	CountPendingURLs(ctx context.Context, chatbotID string) (int, error)

	// UpdatePendingURLStatus updates the status of multiple pending URLs.
	UpdatePendingURLStatus(ctx context.Context, chatbotID string, urlIDs []string, status string) (int, error)

	// GetPendingURLsByIDs returns pending URLs by their IDs.
	GetPendingURLsByIDs(ctx context.Context, chatbotID string, urlIDs []string) ([]models.PendingURL, error)

	// DeletePendingURLsByChatbot clears all pending URLs for a chatbot.
	DeletePendingURLsByChatbot(ctx context.Context, chatbotID string) (int, error)
}

// OrganizationFilter contains optional filters for listing organizations in admin views.
type OrganizationFilter struct {
	Name   *string
	PlanID *string
}

// OrganizationRepository defines the interface for organization data access operations.
type OrganizationRepository interface {
	GetByID(ctx context.Context, id string) (*models.Organization, error)
	AdminList(ctx context.Context, filter OrganizationFilter, limit, offset int) ([]*models.Organization, int, error)
	GetPlatformOverviewStats(ctx context.Context) (*PlatformOverviewStats, error)
	GetTotalOrganizations(ctx context.Context) (int, error)
	GetTotalChatbots(ctx context.Context) (int, error)
}

// PlatformOverviewStats represents aggregated platform statistics for admin views.
type PlatformOverviewStats struct {
	TotalUsers         int `json:"total_users"`
	TotalOrganizations int `json:"total_organizations"`
	TotalChatbots      int `json:"total_chatbots"`
	TotalMessages      int `json:"total_messages"`
}

// AnalyticsPoint represents a single data point in analytics time series.
type AnalyticsPoint struct {
	Date          string `json:"date"`
	Messages      int    `json:"messages"`
	Conversations int    `json:"conversations"`
	Tokens        int    `json:"tokens"`
	ThumbsUp      int    `json:"thumbs_up"`
	ThumbsDown    int    `json:"thumbs_down"`
	Handoffs      int    `json:"handoffs"`
}

// SourceUsageStat represents usage statistics for a data source.
type SourceUsageStat struct {
	SourceID         string `json:"source_id"`
	SourceType       string `json:"source_type"`
	SourceURL        string `json:"source_url,omitempty"`
	OriginalFilename string `json:"original_filename,omitempty"`
	MessageCount     int    `json:"message_count"`
}

// AuditLogEntry represents an entry in the admin audit log.
type AuditLogEntry struct {
	ID          string         `json:"id"`
	AdminUserID string         `json:"admin_user_id"`
	Action      string         `json:"action"`
	TargetType  string         `json:"target_type"`
	TargetID    *string        `json:"target_id"`
	Details     map[string]any `json:"details"`
	IPAddress   string         `json:"ip_address"`
	UserAgent   string         `json:"user_agent"`
	CreatedAt   time.Time      `json:"created_at"`
}

// AuditFilter contains optional filters for listing audit logs.
type AuditFilter struct {
	AdminUserID *string
	Action      *string
	TargetType  *string
	TargetID    *string
	StartDate   *time.Time
	EndDate     *time.Time
}

// AdminSource represents a data source for admin views with additional metadata.
type AdminSource struct {
	ID               string     `json:"id"`
	ChatbotID        string     `json:"chatbot_id"`
	ChatbotName      string     `json:"chatbot_name"`
	OrganizationName *string    `json:"organization_name,omitempty"`
	OwnerEmail       string     `json:"owner_email"`
	SourceType       string     `json:"source_type"`
	SourceURL        *string    `json:"source_url,omitempty"`
	OriginalFilename *string    `json:"original_filename,omitempty"`
	Status           string     `json:"status"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	ChunkCount       int        `json:"chunk_count"`
	SizeBytes        *int64     `json:"size_bytes,omitempty"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	CreatedAt        string     `json:"created_at"`
}

// AdminSourceFilter contains optional filters for listing sources in admin views.
type AdminSourceFilter struct {
	ChatbotID  *string
	SourceType *string
	Status     *string
	OwnerID    *string
}

// SourceStats represents aggregated statistics for data sources.
type SourceStats struct {
	StatusCounts map[string]int `json:"status_counts"`
}

// ErrorLogEntry represents an entry in the error log.
type ErrorLogEntry struct {
	ID             string    `json:"id"`
	ErrorType      string    `json:"error_type"`
	Message        string    `json:"message"`
	StackTrace     string    `json:"stack_trace,omitempty"`
	RequestPath    string    `json:"request_path,omitempty"`
	RequestMethod  string    `json:"request_method,omitempty"`
	UserID         *string   `json:"user_id,omitempty"`
	ChatbotID      *string   `json:"chatbot_id,omitempty"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	Severity       string    `json:"severity"`
	Context        []byte    `json:"context,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// ErrorStats represents aggregated error statistics.
type ErrorStats struct {
	SeverityCounts map[string]int `json:"severity_counts"`
}

// AdminRepository defines the interface for admin-only data access operations.
// These methods are for platform administrators and have elevated access privileges.
type AdminRepository interface {
	// Audit Log Operations

	// InsertAuditLog persists a new audit log entry.
	InsertAuditLog(ctx context.Context, entry AuditLogEntry) error

	// ListAuditLogs returns a paginated list of audit logs with optional filtering.
	ListAuditLogs(ctx context.Context, filter AuditFilter, limit, offset int) ([]AuditLogEntry, int, error)

	// Source Management

	// AdminListSources returns a paginated list of all data sources with metadata.
	AdminListSources(ctx context.Context, filter AdminSourceFilter, limit, offset int) ([]AdminSource, int, error)

	// AdminGetSourceByID retrieves a single source by ID with all admin-visible details.
	AdminGetSourceByID(ctx context.Context, id string) (*AdminSource, error)

	// AdminGetSourceStats returns aggregated statistics for data sources.
	AdminGetSourceStats(ctx context.Context) (*SourceStats, error)

	// AdminReprocessSource resets a source to pending status for reprocessing.
	AdminReprocessSource(ctx context.Context, id string) error

	// Error Logs

	// ListErrorLogs returns a paginated list of error logs with optional severity filtering.
	ListErrorLogs(ctx context.Context, severity string, limit, offset int) ([]ErrorLogEntry, int, error)

	// GetErrorLogByID retrieves a single error log entry by ID.
	GetErrorLogByID(ctx context.Context, id string) (*ErrorLogEntry, error)

	// GetErrorStats returns aggregated error statistics for the last 24 hours.
	GetErrorStats(ctx context.Context) (*ErrorStats, error)
}
