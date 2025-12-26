package policy

// Model represents a typed model identifier.
// This eliminates magic strings for model names scattered across the codebase.
// The actual model metadata (context windows, capabilities, etc.) is stored in the ai_models database table.
type Model string

// Model constants define the currently supported AI models in the system.
// These use the bare model names (e.g., "gpt-4o-mini", not "openai/gpt-4o-mini").
// The provider prefix is added by the system when needed for API calls to OpenRouter.
const (
	// Chat/Completion models (via OpenRouter)
	ModelGPT4o     Model = "gpt-4o"
	ModelGPT4oMini Model = "gpt-4o-mini"
	ModelGPT5      Model = "gpt-5"
	
	// Embedding models (via OpenAI directly)
	ModelEmbeddingSmall Model = "text-embedding-3-small"
)

// String returns the string representation of the model.
func (m Model) String() string {
	return string(m)
}

// IsValid checks if the model is one of the recognized model types.
// Note: This only validates known constants, not all possible models in the database.
func (m Model) IsValid() bool {
	switch m {
	case ModelGPT4o, ModelGPT4oMini, ModelGPT5, ModelEmbeddingSmall:
		return true
	default:
		return false
	}
}

// DefaultChatModel returns the default model for chat completions.
// This should match the default in pkg/config/models.go and database migrations.
func DefaultChatModel() Model {
	return ModelGPT4oMini
}

// DefaultEmbeddingModel returns the default model for embeddings.
func DefaultEmbeddingModel() Model {
	return ModelEmbeddingSmall
}
