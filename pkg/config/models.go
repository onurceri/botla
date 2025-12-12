package config

// Model name constants - use these instead of hardcoded strings throughout the codebase.
// These are the standard model identifiers used by the respective providers.

// OpenAI Models (bare model IDs, used for direct OpenAI API and validation)
const (
	ModelGPT4oMini  = "gpt-4o-mini"
	ModelGPT4o      = "gpt-4o"
	ModelGPT4Turbo  = "gpt-4-turbo"
	ModelGPT4       = "gpt-4"
	ModelGPT35Turbo = "gpt-3.5-turbo"
	ModelO1         = "o1"
	ModelO1Mini     = "o1-mini"
	ModelO1Preview  = "o1-preview"
	ModelO3Mini     = "o3-mini"
)

// Anthropic Models (bare IDs; routed via OpenRouter using provider prefix)
const (
	ModelClaude35Sonnet = "claude-3-5-sonnet-20241022"
)

// Google Models (bare IDs; routed via OpenRouter using provider prefix)
const (
	ModelGemini15Flash = "gemini-1.5-flash"
)

// Embedding models (OpenAI-compatible)
const (
	ModelEmbeddingSmall = "text-embedding-3-small"
)

// OpenRouter-specific model identifiers (provider/model format)
const (
	ModelOpenRouterGPT4oMini = "openai/gpt-4o-mini"
)

// DefaultOpenAIModel returns the default model for direct OpenAI API usage
func DefaultOpenAIModel() string {
	return ModelGPT4oMini
}

// DefaultOpenRouterModel returns the default model for OpenRouter usage
func DefaultOpenRouterModel() string {
	return ModelOpenRouterGPT4oMini
}
