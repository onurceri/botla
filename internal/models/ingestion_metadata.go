package models

// IngestionMetadata contains metadata extracted during content ingestion.
type IngestionMetadata struct {
	CapabilitySummary  string   `json:"capability_summary"`
	SuggestedQuestions []string `json:"suggested_questions"`
}
