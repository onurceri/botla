package models

import "time"

type Conversation struct {
    ID          string     `json:"id"`
    ChatbotID   string     `json:"chatbot_id"`
    SessionID   *string    `json:"session_id,omitempty"`
    MessageCount int       `json:"message_count"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

