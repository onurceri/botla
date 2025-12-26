package policy

// This package provides commonly referenced limit values and constants.
// IMPORTANT: The actual plan limits are stored in the database (plans.config JSONB field).
// These constants are for reference and validation only, NOT as the source of truth.

// Common token limit values referenced in the codebase
const (
	// Token limits per plan (for reference - actual values in database)
	TokenLimitFree  int64 = 100_000
	TokenLimitPro   int64 = 1_000_000
	TokenLimitUltra int64 = 10_000_000
)

// Common chatbot count limits (for reference - actual values in database)
const (
	MaxChatbotsFree  = 1
	MaxChatbotsPro   = 5
	MaxChatbotsUltra = 100
)

// Note: Plan-specific limits (AllowedModels, RAG settings, etc.) are configured
// in the database and should be retrieved via the Plan.Config field.
// Do not hardcode plan limits in application code.

