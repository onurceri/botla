package models

import "time"

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	TokensUsed     int       `json:"tokens_used"`
	ThumbsUp       *bool           `json:"thumbs_up,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	Sources        []MessageSource `json:"sources,omitempty"`
}

type MessageSource struct {
	ID             string    `json:"id"`
	MessageID      string    `json:"message_id"`
	SourceID       string    `json:"source_id"`
	ChunkIndex     int       `json:"chunk_index"`
	RelevanceScore float64   `json:"relevance_score"`
	CreatedAt      time.Time `json:"created_at"`
}
