package models

import (
	"time"

	"github.com/google/uuid"
)

// AIModel represents a record in the ai_models table
// The table separates model identification into:
// - ModelName: bare model name (e.g., "gpt-4o-mini") - use for comparisons and display
// - APIModelID: full OpenRouter format (e.g., "openai/gpt-4o-mini") - use for API calls
// This eliminates runtime parsing of provider prefixes (/, :)
type AIModel struct {
	ID         uuid.UUID `json:"id"`
	Provider   string    `json:"provider"`      // Provider name (e.g., "openai", "anthropic")
	ModelName  string    `json:"model_name"`    // Bare model name (e.g., "gpt-4o-mini")
	APIModelID string    `json:"api_model_id"`  // Full API identifier (e.g., "openai/gpt-4o-mini")
	Name       string    `json:"name"`          // Display name (e.g., "GPT-4o Mini")
	MaxTokens  int       `json:"max_tokens"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ToModelInfo converts DB model to API response model
func (m *AIModel) ToModelInfo() ModelInfo {
	return ModelInfo{
		ID:                m.ModelName,  // Use bare model name for API responses
		Name:              m.Name,
		Provider:          m.Provider,
		MaxTokens:         m.MaxTokens,
		SupportedFeatures: []string{},
	}
}

// GetAPIModelID returns the full API identifier for LLM calls
// This is the format expected by OpenRouter (e.g., "openai/gpt-4o-mini")
func (m *AIModel) GetAPIModelID() string {
	return m.APIModelID
}
