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
	UserEmail      *string    `json:"user_email,omitempty"`
	ContactedAt    *time.Time `json:"contacted_at,omitempty"` // when the visitor submitted UserEmail
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

// HandoffStatus constants
const (
	HandoffStatusPending  = "pending"
	HandoffStatusAssigned = "assigned"
	HandoffStatusResolved = "resolved"
)

// IsValidHandoffStatus reports whether status is one of the known handoff statuses.
func IsValidHandoffStatus(status string) bool {
	switch status {
	case HandoffStatusPending, HandoffStatusAssigned, HandoffStatusResolved:
		return true
	}
	return false
}

// HandoffListFilter narrows and pages a chatbot's handoff request list.
// An empty Status matches all statuses.
type HandoffListFilter struct {
	Status string
	Limit  int
	Offset int
}
