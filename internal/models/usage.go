package models

// Usage represents user usage statistics
type Usage struct {
	FilesCount               int `json:"files_count"`
	MaxFilesCountInOneBot    int `json:"max_files_count_in_one_bot"`
	StorageUsedMB            int `json:"storage_used_mb"`
	URLsCount                int `json:"urls_count"`
	MaxURLsCountInOneBot     int `json:"max_urls_count_in_one_bot"`
	TokensUsed               int `json:"tokens_used"`
	IngestionsUsed           int `json:"ingestions_used"`
	IngestionEmbeddingTokens int `json:"ingestion_embedding_tokens"`
	RefreshCount             int `json:"refresh_count"`
}
