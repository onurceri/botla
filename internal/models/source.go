package models

import "time"

type DataSource struct {
	ID               string     `json:"id"`
	ChatbotID        string     `json:"chatbot_id"`
	SourceType       string     `json:"source_type"`
	SourceURL        *string    `json:"source_url,omitempty"`
	FilePath         *string    `json:"file_path,omitempty"`
	OriginalFilename *string    `json:"original_filename,omitempty"`
	Status           string     `json:"status"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	ChunkCount       int        `json:"chunk_count"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}
