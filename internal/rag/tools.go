package rag

import (
	"encoding/json"
	"regexp"

	"github.com/onurceri/botla-co/internal/models"
)

var toolNameRE = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

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

func IsValidToolName(name string) bool {
	return toolNameRE.MatchString(name)
}

func normalizeToolParameters(raw *json.RawMessage) json.RawMessage {
	if raw == nil || len(*raw) == 0 {
		return json.RawMessage(`{"type": "object", "properties": {}}`)
	}

	var m map[string]any
	if err := json.Unmarshal(*raw, &m); err != nil {
		return json.RawMessage(`{"type": "object", "properties": {}}`)
	}
	if m == nil {
		return json.RawMessage(`{"type": "object", "properties": {}}`)
	}
	if t, ok := m["type"].(string); !ok || t != "object" {
		m["type"] = "object"
	}
	if _, ok := m["properties"]; !ok {
		m["properties"] = map[string]any{}
	}

	if b, err := json.Marshal(m); err == nil {
		return b
	}
	return json.RawMessage(`{"type": "object", "properties": {}}`)
}

// ConvertActionsToTools converts ChatbotActions to OpenAI tool format
func ConvertActionsToTools(actions []*models.ChatbotAction) []Tool {
	var tools []Tool
	for _, action := range actions {
		if !action.Enabled {
			continue
		}
		// Use the LLM-generated tool_name, skip if not available
		if action.ToolName == nil || *action.ToolName == "" {
			continue
		}
		toolName := *action.ToolName
		if !IsValidToolName(toolName) {
			continue
		}
		desc := ""
		if action.Description != nil {
			desc = *action.Description
		}
		params := normalizeToolParameters(action.Parameters)

		tools = append(tools, Tool{
			Type: "function",
			Function: ToolFunction{
				Name:        toolName, // Use LLM-generated tool_name
				Description: desc,
				Parameters:  params,
			},
		})
	}
	return tools
}

// BuiltinToolOptions configures which built-in tools to include
type BuiltinToolOptions struct {
	IncludeListSources bool // Include list_sources tool (default: true)
	IncludeHandoff     bool // Include request_human_handoff tool
}

// DefaultBuiltinToolOptions returns the default options with list_sources enabled
func DefaultBuiltinToolOptions() BuiltinToolOptions {
	return BuiltinToolOptions{
		IncludeListSources: true,
		IncludeHandoff:     false,
	}
}

// GetBuiltinToolsWithOptions returns built-in tools based on options struct
func GetBuiltinToolsWithOptions(options BuiltinToolOptions) []Tool {
	var tools []Tool

	if options.IncludeListSources {
		tools = append(tools, Tool{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_sources",
				Description: "Lists the available knowledge sources and their capabilities",
				Parameters:  normalizeToolParameters(nil),
			},
		})
	}

	if options.IncludeHandoff {
		tools = append(tools, Tool{
			Type: "function",
			Function: ToolFunction{
				Name:        "request_human_handoff",
				Description: "Request to transfer the conversation to a human support agent. Use this when you cannot help the user or when they explicitly ask for a human.",
				Parameters:  normalizeToolParameters(nil),
			},
		})
	}

	return tools
}
