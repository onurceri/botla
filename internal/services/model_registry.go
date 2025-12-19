package services

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/models"
)

// ModelService handles logic related to AI models
type ModelService struct {
	DB *sql.DB
}

// NewModelService creates a new ModelService
func NewModelService(db *sql.DB) *ModelService {
	return &ModelService{DB: db}
}

// GetAvailableModels returns the list of models available for the given plan configuration
func (s *ModelService) GetAvailableModels(ctx context.Context, allowedModels []string) ([]models.ModelInfo, error) {
	// Fetch all active models from DB
	rows, err := s.DB.QueryContext(ctx, "SELECT id, provider, model_id, name, max_tokens, is_active FROM ai_models WHERE is_active = true")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var allModels []models.ModelInfo
	for rows.Next() {
		var m models.AIModel
		if err := rows.Scan(&m.ID, &m.Provider, &m.ModelID, &m.Name, &m.MaxTokens, &m.IsActive); err != nil {
			continue
		}
		allModels = append(allModels, m.ToModelInfo())
	}

	// If no allowed models are specified, return all (fallback behavior)
	if len(allowedModels) == 0 {
		return allModels, nil
	}

	var available []models.ModelInfo
	allowedSet := make(map[string]bool)
	for _, m := range allowedModels {
		allowedSet[m] = true
	}

	for _, model := range allModels {
		// Check exact match on ModelID (e.g. openai/gpt-4o-mini)
		if allowedSet[model.ID] {
			available = append(available, model)
		}
	}

	return available, nil
}
