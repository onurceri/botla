package rag

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/logger"
)

type ToolExecutor struct {
	DB  *sql.DB
	Log *logger.Logger
}

type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Result     string `json:"result"` // JSON string
	Error      string `json:"error,omitempty"`
}

// Execute executes a tool call and returns the result
func (e *ToolExecutor) Execute(ctx context.Context, toolCall ToolCall, action *models.ChatbotAction, chatbotID, conversationID string) (*ToolResult, error) {
	if action == nil {
		// Built-in tools
		switch toolCall.Function.Name {
		case "list_sources":
			return e.executeBuiltin(toolCall)
		case "request_human_handoff":
			return e.executeHandoff(ctx, toolCall, chatbotID, conversationID)
		default:
			return nil, fmt.Errorf("unknown action: %s", toolCall.Function.Name)
		}
	}

	switch action.ActionType {
	case models.ActionTypeBuiltin:
		return e.executeBuiltin(toolCall)
	case models.ActionTypeHTTP:
		return e.executeHTTP(ctx, toolCall, action)
	case models.ActionTypeZapier:
		return e.executeZapier(ctx, toolCall, action)
	default:
		return nil, fmt.Errorf("unknown action type: %s", action.ActionType)
	}
}

func (e *ToolExecutor) executeBuiltin(toolCall ToolCall) (*ToolResult, error) {
	switch toolCall.Function.Name {
	case "list_sources":
		// Return capability summaries
		// In a real implementation, this would query the DB for sources
		return &ToolResult{
			ToolCallID: toolCall.ID,
			Result:     `{"sources": "Knowledge base contains documents and website content loaded into the chatbot."}`,
		}, nil
	default:
		return nil, fmt.Errorf("unknown builtin tool: %s", toolCall.Function.Name)
	}
}

func (e *ToolExecutor) executeHandoff(ctx context.Context, toolCall ToolCall, chatbotID, conversationID string) (*ToolResult, error) {
	// Check for existing active request
	exists, err := db.HasActiveHandoffRequest(ctx, e.DB, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing handoff: %w", err)
	}
	if exists {
		return &ToolResult{
			ToolCallID: toolCall.ID,
			Result:     `{"status": "error", "message": "A handoff request is already active for this conversation."}`,
		}, nil
	}

	// Create new request
	req := &models.HandoffRequest{
		ChatbotID:      chatbotID,
		ConversationID: conversationID,
		Status:         models.HandoffStatusPending,
		Notes:          nil, // No notes from tool for now, or could parse from args
	}

	requestID, err := db.CreateHandoffRequest(ctx, e.DB, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create handoff request: %w", err)
	}

	return &ToolResult{
		ToolCallID: toolCall.ID,
		Result:     fmt.Sprintf(`{"status": "handoff_requested", "request_id": "%s"}`, requestID),
	}, nil
}

func (e *ToolExecutor) executeHTTP(ctx context.Context, toolCall ToolCall, action *models.ChatbotAction) (*ToolResult, error) {
	var config models.HTTPActionConfig
	if action.Config == nil {
		return nil, fmt.Errorf("http config is missing")
	}
	if err := json.Unmarshal(*action.Config, &config); err != nil {
		return nil, fmt.Errorf("invalid http config: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	method := config.Method
	if method == "" {
		method = "POST"
	}

	var body io.Reader
	if toolCall.Function.Arguments != "" {
		body = bytes.NewBufferString(toolCall.Function.Arguments)
	}

	req, err := http.NewRequestWithContext(ctx, method, config.URL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	// Auth
	switch config.AuthType {
	case "bearer":
		// Simplified: assuming token is in AuthConfig
		var authCfg map[string]string
		_ = json.Unmarshal(config.AuthConfig, &authCfg)
		if token, ok := authCfg["token"]; ok {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "api_key":
		var authCfg map[string]string
		_ = json.Unmarshal(config.AuthConfig, &authCfg)
		keyName := authCfg["key"]
		val := authCfg["value"]
		in := authCfg["in"] // header or query
		if in == "query" {
			q := req.URL.Query()
			q.Add(keyName, val)
			req.URL.RawQuery = q.Encode()
		} else {
			req.Header.Set(keyName, val)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		ToolCallID: toolCall.ID,
		Result:     string(respBody),
	}, nil
}

func (e *ToolExecutor) executeZapier(ctx context.Context, toolCall ToolCall, action *models.ChatbotAction) (*ToolResult, error) {
	var config models.ZapierActionConfig
	if action.Config == nil {
		return nil, fmt.Errorf("zapier config is missing")
	}
	if err := json.Unmarshal(*action.Config, &config); err != nil {
		return nil, fmt.Errorf("invalid zapier config: %w", err)
	}

	if config.WebhookURL == "" {
		return nil, fmt.Errorf("zapier webhook url is empty")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var body io.Reader
	if toolCall.Function.Arguments != "" {
		body = bytes.NewBufferString(toolCall.Function.Arguments)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		ToolCallID: toolCall.ID,
		Result:     string(respBody),
	}, nil
}
