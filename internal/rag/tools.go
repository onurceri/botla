package rag

import (
	"encoding/json"

	"github.com/onurceri/botla-co/internal/models"
)

type ChatMessage struct {
	Role       string     `json:"role"`
	Content    *string    `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// OpenAI Function Calling tool format
type Tool struct {
	Type     string       `json:"type"` // "function"
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string
	} `json:"function"`
}

// ConvertActionsToTools converts ChatbotActions to OpenAI tool format
func ConvertActionsToTools(actions []*models.ChatbotAction) []Tool {
	var tools []Tool
	for _, action := range actions {
		if !action.Enabled {
			continue
		}
		desc := ""
		if action.Description != nil {
			desc = *action.Description
		}
		var params json.RawMessage
		if action.Parameters != nil {
			params = *action.Parameters
		} else {
			params = json.RawMessage(`{"type": "object", "properties": {}}`)
		}

		tools = append(tools, Tool{
			Type: "function",
			Function: ToolFunction{
				Name:        action.Name,
				Description: desc,
				Parameters:  params,
			},
		})
	}
	return tools
}

// Built-in tools
func GetBuiltinTools() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_sources",
				Description: "Lists the available knowledge sources and their capabilities",
				Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "request_human_handoff",
				Description: "Request to transfer the conversation to a human support agent. Use this when you cannot help the user or when they explicitly ask for a human.",
				Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
			},
		},
	}
}
