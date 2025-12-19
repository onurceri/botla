package models

import (
	"time"

	"github.com/google/uuid"
)

// AIModel represents a record in the ai_models table
type AIModel struct {
	ID        uuid.UUID `json:"id"`
	Provider  string    `json:"provider"`
	ModelID   string    `json:"model_id"`
	Name      string    `json:"name"`
	MaxTokens int       `json:"max_tokens"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToModelInfo converts DB model to API response model
func (m *AIModel) ToModelInfo() ModelInfo {
	return ModelInfo{
		ID:        m.ModelID, // Use the string ID (e.g., openai/gpt-4o) for API
		Name:      m.Name,
		Provider:  m.Provider,
		MaxTokens: m.MaxTokens,
		// SupportedFeatures can be added to DB if needed, for now empty or default
		SupportedFeatures: []string{},
	}
}
