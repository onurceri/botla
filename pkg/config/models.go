package config

// DefaultModelName is the fallback model when DEFAULT_CHATBOT_MODEL env var is not set.
// This should match a model_name in the ai_models table.
const DefaultModelName = "gpt-4o-mini"

// ModelEmbeddingSmall is the embedding model used by the system.
// This is a system constant, not user-configurable.
const ModelEmbeddingSmall = "text-embedding-3-small"
