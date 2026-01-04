package config

import "github.com/onurceri/botla-app/pkg/policy"

// DefaultModelName is the fallback model when DEFAULT_CHATBOT_MODEL env var is not set.
// This should match a model_name in the ai_models table.
// Uses the policy package to ensure consistency across the codebase.
var DefaultModelName = policy.DefaultChatModel().String()

// ModelEmbeddingSmall is the embedding model used by the system.
// This is a system constant, not user-configurable.
// Uses the policy package to ensure consistency across the codebase.
var ModelEmbeddingSmall = policy.DefaultEmbeddingModel().String()
