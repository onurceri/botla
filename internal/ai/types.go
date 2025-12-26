package ai

import "time"

// VectorPayload represents metadata stored alongside embeddings
type VectorPayload struct {
	ChatbotID    string    `json:"chatbot_id"`
	SourceID     string    `json:"source_id"`
	ChunkIndex   int       `json:"chunk_index"`
	OriginalText string    `json:"original_text"`
	SourceType   string    `json:"source_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// SearchFilter defines criteria for vector searches
type SearchFilter struct {
	ChatbotID string
	SourceID  string
}

// DeleteFilter defines criteria for deleting vectors
type DeleteFilter struct {
	SourceID string
}

// SearchResult represents a single vector search result
type SearchResult struct {
	ID      interface{}
	Score   float64
	Payload VectorPayload
}
