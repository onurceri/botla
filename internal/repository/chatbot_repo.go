// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
)

// PostgresChatbotRepo implements ChatbotRepository using PostgreSQL.
// It delegates to the existing db package functions to provide a gradual migration path.
type PostgresChatbotRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresChatbotRepo implements ChatbotRepository.
var _ ChatbotRepository = (*PostgresChatbotRepo)(nil)

// NewPostgresChatbotRepo creates a new PostgresChatbotRepo instance.
func NewPostgresChatbotRepo(pool *sql.DB) *PostgresChatbotRepo {
	return &PostgresChatbotRepo{pool: pool}
}

// GetByID retrieves a chatbot by its unique identifier.
// Returns nil, nil if the chatbot is not found.
func (r *PostgresChatbotRepo) GetByID(ctx context.Context, id string) (*models.Chatbot, error) {
	return db.GetChatbotByID(ctx, r.pool, id)
}

// GetByUserID retrieves all non-deleted chatbots for a user.
func (r *PostgresChatbotRepo) GetByUserID(ctx context.Context, userID string) ([]models.Chatbot, error) {
	return db.GetChatbotsByUserID(ctx, r.pool, userID)
}

// GetByWorkspace retrieves all non-deleted chatbots for a workspace.
func (r *PostgresChatbotRepo) GetByWorkspace(ctx context.Context, workspaceID string) ([]models.Chatbot, error) {
	return db.GetChatbotsByWorkspace(ctx, r.pool, workspaceID)
}

// Create persists a new chatbot and returns its generated ID.
func (r *PostgresChatbotRepo) Create(ctx context.Context, bot *models.Chatbot) (string, error) {
	return db.CreateChatbot(ctx, r.pool, bot)
}

// Update modifies an existing chatbot's fields.
func (r *PostgresChatbotRepo) Update(ctx context.Context, bot *models.Chatbot) error {
	return db.UpdateChatbot(ctx, r.pool, bot)
}

// SoftDelete marks a chatbot as deleted and returns the IDs of associated sources.
// This allows cleanup of related resources (e.g., vectors in Qdrant).
func (r *PostgresChatbotRepo) SoftDelete(ctx context.Context, id, userID string) ([]string, error) {
	return db.SoftDeleteChatbot(ctx, r.pool, id, userID)
}

// CountByUserID returns the number of active chatbots for a user.
func (r *PostgresChatbotRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	return db.CountChatbotsByUserID(ctx, r.pool, userID)
}

// CountByWorkspace returns the number of active chatbots for a workspace.
func (r *PostgresChatbotRepo) CountByWorkspace(ctx context.Context, workspaceID string) (int, error) {
	return db.CountChatbotsByWorkspace(ctx, r.pool, workspaceID)
}

// UpdateSuggestedQuestions updates only the AI-generated suggestions.
func (r *PostgresChatbotRepo) UpdateSuggestedQuestions(ctx context.Context, id string, suggestions []string) error {
	return db.UpdateChatbotSuggestedQuestions(ctx, r.pool, id, suggestions)
}
