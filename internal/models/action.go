package models

import (
	"encoding/json"
	"time"
)

type ActionType string

const (
	ActionTypeBuiltin ActionType = "builtin"
	ActionTypeHTTP    ActionType = "http"
	ActionTypeZapier  ActionType = "zapier"
)

type ChatbotAction struct {
	ID          string           `json:"id"`
	ChatbotID   string           `json:"chatbot_id"`
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	ActionType  ActionType       `json:"action_type"`
	Config      *json.RawMessage `json:"config"`
	Parameters  *json.RawMessage `json:"parameters"` // JSON Schema
	ToolName    *string          `json:"tool_name"`  // LLM-generated API-compatible identifier
	Enabled     bool             `json:"enabled"`
	Version     int              `json:"version"` // For optimistic locking
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   *time.Time       `json:"updated_at"`
}

// HTTP Action config
type HTTPActionConfig struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	AuthType   string            `json:"auth_type"` // none, bearer, api_key
	AuthConfig json.RawMessage   `json:"auth_config"`
}

// Zapier Action config
type ZapierActionConfig struct {
	WebhookURL string `json:"webhook_url"`
}

type ActionExecutionLog struct {
	ID              string           `json:"id"`
	ChatbotID       string           `json:"chatbot_id"`
	ActionID        string           `json:"action_id"`
	ConversationID  *string          `json:"conversation_id"`
	MessageID       *string          `json:"message_id"`
	Status          string           `json:"status"` // "success", "failure"
	RequestPayload  *json.RawMessage `json:"request_payload"`
	ResponsePayload *json.RawMessage `json:"response_payload"`
	ErrorMessage    *string          `json:"error_message,omitempty"`
	DurationMs      int              `json:"duration_ms"`
	CreatedAt       time.Time        `json:"created_at"`
}
