package models

// ChatRequest represents the input for the chat process
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

// SourceUsed represents a source chunk used in the answer
type SourceUsed struct {
	ChunkIndex int    `json:"chunk_index"`
	SourceType string `json:"source_type"`
}

// ChatResult represents the output of the chat process
type ChatResult struct {
	Response       string       `json:"response"`
	TokensUsed     int          `json:"tokens_used"`
	Sources        []SourceUsed `json:"sources"`
	ConversationID string       `json:"conversation_id"`
	MessageID      string       `json:"message_id"`
	IsNewConv      bool         `json:"is_new_conversation"`
}
