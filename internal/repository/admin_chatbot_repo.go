// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/db"
)

// PostgresAdminChatbotRepo implements AdminChatbotRepository using PostgreSQL.
// It delegates to the existing db package functions for admin chatbot operations.
type PostgresAdminChatbotRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresAdminChatbotRepo implements AdminChatbotRepository.
var _ AdminChatbotRepository = (*PostgresAdminChatbotRepo)(nil)

// NewPostgresAdminChatbotRepo creates a new PostgresAdminChatbotRepo instance.
func NewPostgresAdminChatbotRepo(pool *sql.DB) *PostgresAdminChatbotRepo {
	return &PostgresAdminChatbotRepo{pool: pool}
}

// ListChatbots returns a paginated list of all chatbots with their metadata.
func (r *PostgresAdminChatbotRepo) ListChatbots(ctx context.Context, filter AdminChatbotFilter, limit, offset int) ([]AdminChatbot, int, error) {
	// Convert repository filter to db filter
	dbFilter := db.ChatbotFilter{
		Name:           filter.Name,
		OrganizationID: filter.OrganizationID,
		OwnerID:        filter.OwnerID,
	}

	dbChatbots, total, err := db.AdminListChatbots(ctx, r.pool, dbFilter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Convert db.AdminChatbot to repository.AdminChatbot
	chatbots := make([]AdminChatbot, len(dbChatbots))
	for i, c := range dbChatbots {
		chatbots[i] = toRepoAdminChatbot(c)
	}

	return chatbots, total, nil
}

// GetByID retrieves a single chatbot by ID with all admin-visible details.
func (r *PostgresAdminChatbotRepo) GetByID(ctx context.Context, id string) (*AdminChatbot, error) {
	dbChatbot, err := db.AdminGetChatbot(ctx, r.pool, id)
	if err != nil {
		return nil, err
	}

	result := toRepoAdminChatbot(*dbChatbot)
	return &result, nil
}

// ResetSources resets all sources for a chatbot to pending status for reprocessing.
func (r *PostgresAdminChatbotRepo) ResetSources(ctx context.Context, chatbotID string) (int64, error) {
	return db.AdminResetChatbotSources(ctx, r.pool, chatbotID)
}

// GetSourceIDs returns all pending source IDs for a chatbot for queue processing.
func (r *PostgresAdminChatbotRepo) GetSourceIDs(ctx context.Context, chatbotID string) ([]string, error) {
	return db.AdminGetChatbotSourceIDs(ctx, r.pool, chatbotID)
}

// DeleteVectors resets chunk counts to 0 for all sources (for reindexing).
func (r *PostgresAdminChatbotRepo) DeleteVectors(ctx context.Context, chatbotID string) error {
	return db.AdminDeleteChatbotVectors(ctx, r.pool, chatbotID)
}

// toRepoAdminChatbot converts db.AdminChatbot to repository.AdminChatbot.
func toRepoAdminChatbot(c db.AdminChatbot) AdminChatbot {
	result := AdminChatbot{
		ID:             c.ID,
		Name:           c.Name,
		OwnerID:        c.OwnerID,
		OwnerEmail:     c.OwnerEmail,
		SourceCount:    c.SourceCount,
		MessageCount:   c.MessageCount,
		CustomBranding: c.CustomBranding,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
	if c.WorkspaceID != nil {
		result.WorkspaceID = *c.WorkspaceID
	}
	if c.OrganizationID != nil {
		result.OrganizationID = *c.OrganizationID
	}
	if c.OrganizationName != nil {
		result.OrganizationName = *c.OrganizationName
	}
	return result
}
