package models

import "time"

// PendingURL represents a discovered URL awaiting user approval
type PendingURL struct {
	ID           string    `json:"id"`
	ChatbotID    string    `json:"chatbot_id"`
	SourceID     *string   `json:"source_id,omitempty"`
	URL          string    `json:"url"`
	DiscoveredAt time.Time `json:"discovered_at"`
	Status       string    `json:"status"` // pending, selected, rejected
}
