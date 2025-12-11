package models

import "time"

type DataSource struct {
    ID                string     `json:"id"`
    ChatbotID         string     `json:"chatbot_id"`
    SourceType        string     `json:"source_type"`
    SourceURL         *string    `json:"source_url,omitempty"`
    FilePath          *string    `json:"file_path,omitempty"`
    OriginalFilename  *string    `json:"original_filename,omitempty"`
    Status            string     `json:"status"`
    ErrorMessage      *string    `json:"error_message,omitempty"`
    ChunkCount        int        `json:"chunk_count"`
    ProcessedAt       *time.Time `json:"processed_at,omitempty"`
    CreatedAt         time.Time  `json:"created_at"`
    Hash              *string    `json:"hash,omitempty"`
    DeletedAt         *time.Time `json:"deleted_at,omitempty"`
    SizeBytes         int64      `json:"size_bytes"`
    LastRefreshedAt   *time.Time `json:"last_refreshed_at,omitempty"`
    IsDiscovered      bool       `json:"is_discovered"`
    CapabilitySummary *string    `json:"capability_summary,omitempty"`
}

