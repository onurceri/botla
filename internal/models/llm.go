package models

// CompletionParams defines the parameters for a chat completion request
type CompletionParams struct {
	SystemPrompt string
	Context      string
	UserMessage  string
	Model        string
	Temperature  float32
	MaxTokens    int
}

// CompletionResult defines the result of a chat completion request
type CompletionResult struct {
	Content     string
	UsageTokens int
}

// ModelInfo defines metadata about an LLM model
type ModelInfo struct {
	Name              string
	Provider          string
	MaxTokens         int
	SupportedFeatures []string
}
