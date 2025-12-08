package models

import "time"

// HandoffType represents the type of handoff mechanism
type HandoffType string

const (
	HandoffTypeEmail HandoffType = "email"
)

// HandoffConfig contains configuration for human handoff
type HandoffConfig struct {
	EmailTo      string `json:"email_to,omitempty"`
	EmailSubject string `json:"email_subject,omitempty"`
}

// HandoffRequest represents a request to transfer conversation to human
type HandoffRequest struct {
	ID             string     `json:"id"`
	ChatbotID      string     `json:"chatbot_id"`
	ConversationID string     `json:"conversation_id"`
	Status         string     `json:"status"` // pending, assigned, resolved
	AssignedTo     *string    `json:"assigned_to,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

// HandoffStatus constants
const (
	HandoffStatusPending  = "pending"
	HandoffStatusAssigned = "assigned"
	HandoffStatusResolved = "resolved"
)
