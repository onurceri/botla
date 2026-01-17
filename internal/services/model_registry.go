package services

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
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
// allowedModels should contain bare model names (e.g., "gpt-4o-mini", not "openai/gpt-4o-mini")
func (s *ModelService) GetAvailableModels(ctx context.Context, allowedModels []string) ([]models.ModelInfo, error) {
	// Fetch all active models from DB
	rows, err := s.DB.QueryContext(ctx, "SELECT id, provider, model_name, api_model_id, name, max_tokens, is_active FROM ai_models WHERE is_active = true")
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query models")
	}
	defer func() { _ = rows.Close() }()

	var allModels []models.ModelInfo
	for rows.Next() {
		var m models.AIModel
		if err := rows.Scan(&m.ID, &m.Provider, &m.ModelName, &m.APIModelID, &m.Name, &m.MaxTokens, &m.IsActive); err != nil {
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
		// Check exact match on model name (e.g., "gpt-4o-mini")
		// No more need to parse provider prefixes!
		if allowedSet[model.ID] {
			available = append(available, model)
		}
	}

	return available, nil
}

// GetAPIModelID resolves a bare model name to its full OpenRouter API model ID
// e.g., "gpt-4o-mini" -> "openai/gpt-4o-mini"
// This is the key function that eliminates the need for runtime parsing!
func (s *ModelService) GetAPIModelID(ctx context.Context, modelName string) (string, error) {
	var apiModelID string
	err := s.DB.QueryRowContext(ctx,
		"SELECT api_model_id FROM ai_models WHERE model_name = $1 AND is_active = true",
		modelName,
	).Scan(&apiModelID)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "get api model id")
	}
	return apiModelID, nil
}

// GetModelByName returns the full AIModel record for a given bare model name
func (s *ModelService) GetModelByName(ctx context.Context, modelName string) (*models.AIModel, error) {
	var m models.AIModel
	err := s.DB.QueryRowContext(ctx,
		"SELECT id, provider, model_name, api_model_id, name, max_tokens, is_active FROM ai_models WHERE model_name = $1 AND is_active = true",
		modelName,
	).Scan(&m.ID, &m.Provider, &m.ModelName, &m.APIModelID, &m.Name, &m.MaxTokens, &m.IsActive)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get model by name")
	}
	return &m, nil
}
