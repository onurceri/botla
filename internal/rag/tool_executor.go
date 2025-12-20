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

// FindActionByToolName finds an action by its tool name within a list of actions
func (e *ToolExecutor) FindActionByToolName(actions []*models.ChatbotAction, toolName string) *models.ChatbotAction {
	for _, a := range actions {
		if a.ToolName != nil && *a.ToolName == toolName {
			return a
		}
	}
	return nil
}

// Execute executes a tool call and returns the result
func (e *ToolExecutor) Execute(ctx context.Context, toolCall ToolCall, action *models.ChatbotAction, chatbotID, conversationID string) (*ToolResult, error) {
	start := time.Now()
	var result *ToolResult
	var err error

	defer func() {
		// Only log if we have a valid action (skip builtin tools that are not in DB)
		if action == nil || e.DB == nil {
			return
		}

		duration := int(time.Since(start).Milliseconds())
		status := "success"
		var errorMsg *string
		var responsePayload json.RawMessage

		if err != nil {
			status = "failure"
			msg := err.Error()
			errorMsg = &msg
			// Safely marshal the error message into JSON
			errorJSON, _ := json.Marshal(map[string]string{"error": msg})
			responsePayload = json.RawMessage(errorJSON)
		} else if result != nil {
			// Try to parse result.Result as JSON, otherwise wrap it
			if json.Valid([]byte(result.Result)) {
				responsePayload = json.RawMessage(result.Result)
			} else {
				// Wrap in JSON string if not valid JSON
				b, _ := json.Marshal(result.Result)
				responsePayload = b
			}
		}

		var requestPayload json.RawMessage
		if toolCall.Function.Arguments != "" {
			requestPayload = json.RawMessage(toolCall.Function.Arguments)
		} else {
			requestPayload = json.RawMessage("{}")
		}

		var convID *string
		if conversationID != "" {
			convID = &conversationID
		}

		logEntry := &models.ActionExecutionLog{
			ChatbotID:       chatbotID,
			ActionID:        action.ID,
			ConversationID:  convID,
			Status:          status,
			RequestPayload:  &requestPayload,
			ResponsePayload: &responsePayload,
			ErrorMessage:    errorMsg,
			DurationMs:      duration,
		}

		// Use a detached context for logging
		go func() {
			logCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if logErr := db.CreateActionLog(logCtx, e.DB, logEntry); logErr != nil {
				if e.Log != nil {
					e.Log.Error("failed to create action log", map[string]any{"error": logErr.Error()})
				} else {
					fmt.Printf("failed to create action log: %v\n", logErr)
				}
			}
		}()
	}()

	if action == nil {
		// Built-in tools
		switch toolCall.Function.Name {
		case "list_sources":
			result, err = e.executeBuiltin(ctx, toolCall, chatbotID)
		case "request_human_handoff":
			result, err = e.executeHandoff(ctx, toolCall, chatbotID, conversationID)
		default:
			err = fmt.Errorf("unknown action: %s", toolCall.Function.Name)
		}
		return result, err
	}

	switch action.ActionType {
	case models.ActionTypeBuiltin:
		result, err = e.executeBuiltin(ctx, toolCall, chatbotID)
	case models.ActionTypeHTTP:
		result, err = e.executeHTTP(ctx, toolCall, action)
	case models.ActionTypeZapier:
		result, err = e.executeZapier(ctx, toolCall, action)
	default:
		err = fmt.Errorf("unknown action type: %s", action.ActionType)
	}
	return result, err
}

func (e *ToolExecutor) executeBuiltin(ctx context.Context, toolCall ToolCall, chatbotID string) (*ToolResult, error) {
	switch toolCall.Function.Name {
	case "list_sources":
		if e == nil || e.DB == nil {
			return &ToolResult{
				ToolCallID: toolCall.ID,
				Result:     `{"sources": []}`,
			}, nil
		}
		if chatbotID == "" {
			return &ToolResult{
				ToolCallID: toolCall.ID,
				Result:     `{"sources": []}`,
			}, nil
		}
		sources, err := db.ListSourcesByChatbotID(ctx, e.DB, chatbotID)
		if err != nil {
			return nil, fmt.Errorf("list_sources query failed: %w", err)
		}

		type sourceItem struct {
			ID                  string  `json:"id"`
			SourceType          string  `json:"source_type"`
			Status              string  `json:"status"`
			ChunkCount          int     `json:"chunk_count"`
			SourceURL           *string `json:"source_url,omitempty"`
			FilePath            *string `json:"file_path,omitempty"`
			CapabilitySummary   *string `json:"capability_summary,omitempty"`
			IsDiscovered        bool    `json:"is_discovered"`
			OriginalFilename    *string `json:"original_filename,omitempty"`
			LastRefreshedAtUnix *int64  `json:"last_refreshed_at_unix,omitempty"`
		}

		items := make([]sourceItem, 0, len(sources))
		for i := range sources {
			src := sources[i]
			var refreshedUnix *int64
			if src.LastRefreshedAt != nil {
				u := src.LastRefreshedAt.Unix()
				refreshedUnix = &u
			}
			items = append(items, sourceItem{
				ID:                  src.ID,
				SourceType:          src.SourceType,
				Status:              src.Status,
				ChunkCount:          src.ChunkCount,
				SourceURL:           src.SourceURL,
				FilePath:            src.FilePath,
				CapabilitySummary:   src.CapabilitySummary,
				IsDiscovered:        src.IsDiscovered,
				OriginalFilename:    src.OriginalFilename,
				LastRefreshedAtUnix: refreshedUnix,
			})
		}

		payload := map[string]any{"sources": items}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("list_sources marshal failed: %w", err)
		}
		return &ToolResult{
			ToolCallID: toolCall.ID,
			Result:     string(b),
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
