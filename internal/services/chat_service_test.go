package services

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/langconfig"
)

// =============================================================================
// UNIT TESTS FOR CHAT SERVICE HELPER FUNCTIONS
// =============================================================================

func TestNormalizeLangCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "tr"},
		{"tr", "tr"},
		{"en", "en"},
		{"en-US", "en"},
		{"tr-TR", "tr"},
		{"  ", "tr"},
	}
	for _, tc := range tests {
		result := normalizeLangCode(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeLangCode(%q) = %q; want %q", tc.input, result, tc.expected)
		}
	}
}

func TestCalculateHistoryLimit(t *testing.T) {
	tests := []struct {
		name      string
		maxTokens int
		want      int
	}{
		{"zero tokens defaults to min", 0, 4},
		{"very low tokens defaults to min", 500, 4},
		{"low tokens defaults to min", 1000, 4},
		{"normal tokens 2000", 2000, 5},
		{"normal tokens 4000", 4000, 10},
		{"high tokens 8000", 8000, 20},
		{"very high tokens caps at max", 15000, 20},
		{"extreme tokens caps at max", 50000, 20},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calculateHistoryLimit(tc.maxTokens)
			if got != tc.want {
				t.Errorf("calculateHistoryLimit(%d) = %d, want %d", tc.maxTokens, got, tc.want)
			}
		})
	}
}

func TestParseHandoffRequestID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid request_id",
			input: `{"request_id": "abc-123-def", "status": "ok"}`,
			want:  "abc-123-def",
		},
		{
			name:  "uuid format",
			input: `{"request_id": "550e8400-e29b-41d4-a716-446655440000", "success": true}`,
			want:  "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:  "no request_id field",
			input: `{"status": "ok", "message": "done"}`,
			want:  "",
		},
		{
			name:  "empty request_id",
			input: `{"request_id": "", "status": "ok"}`,
			want:  "",
		},
		{
			name:  "not json",
			input: `plain text response`,
			want:  "",
		},
		{
			name:  "empty string",
			input: ``,
			want:  "",
		},
		{
			name:  "malformed json",
			input: `{"request_id": "abc-123`,
			want:  "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseHandoffRequestID(tc.input)
			if got != tc.want {
				t.Errorf("parseHandoffRequestID() = %q, want %q", got, tc.want)
			}
		})
	}
}

// =============================================================================
// INIT CHAT CONTEXT TESTS
// =============================================================================

func TestInitChatContext_BotName(t *testing.T) {
	service := &ChatService{}

	tests := []struct {
		name           string
		botName        string
		botDisplayName *string
		wantBotName    string
	}{
		{
			name:           "uses bot name when display name is nil",
			botName:        "TestBot",
			botDisplayName: nil,
			wantBotName:    "TestBot",
		},
		{
			name:           "uses bot name when display name is empty",
			botName:        "TestBot",
			botDisplayName: strPtr(""),
			wantBotName:    "TestBot",
		},
		{
			name:           "uses display name when set",
			botName:        "TestBot",
			botDisplayName: strPtr("Friendly Bot"),
			wantBotName:    "Friendly Bot",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bot := &models.Chatbot{
				Name:           tc.botName,
				BotDisplayName: tc.botDisplayName,
				LanguageCode:   "tr",
			}
			req := models.ChatRequest{Message: "test", SessionID: "s1"}
			ragConfig := models.RAGConfig{TopK: 5, MaxContextTokens: 4000}

			cc := service.initChatContext(context.Background(), req, bot, ragConfig, nil)

			if cc.BotName != tc.wantBotName {
				t.Errorf("BotName = %q, want %q", cc.BotName, tc.wantBotName)
			}
		})
	}
}

func TestInitChatContext_ThresholdConfig(t *testing.T) {
	service := &ChatService{}

	t.Run("uses bot threshold config when set", func(t *testing.T) {
		customConfig := &models.ThresholdConfig{
			HighThreshold:   0.8,
			MediumThreshold: 0.5,
			FallbackMode:    "escalate",
		}
		bot := &models.Chatbot{
			Name:            "TestBot",
			LanguageCode:    "tr",
			ThresholdConfig: customConfig,
		}
		req := models.ChatRequest{Message: "test"}
		ragConfig := models.RAGConfig{}

		cc := service.initChatContext(context.Background(), req, bot, ragConfig, nil)

		if cc.ThresholdCfg != customConfig {
			t.Error("should use custom threshold config")
		}
		if cc.ThresholdCfg.HighThreshold != 0.8 {
			t.Errorf("HighThreshold = %f, want 0.8", cc.ThresholdCfg.HighThreshold)
		}
	})

	t.Run("uses default threshold config when nil", func(t *testing.T) {
		bot := &models.Chatbot{
			Name:            "TestBot",
			LanguageCode:    "tr",
			ThresholdConfig: nil,
		}
		req := models.ChatRequest{Message: "test"}
		ragConfig := models.RAGConfig{}

		cc := service.initChatContext(context.Background(), req, bot, ragConfig, nil)

		if cc.ThresholdCfg == nil {
			t.Fatal("ThresholdCfg should not be nil")
		}
		// Default values from models.DefaultThresholdConfig()
		if cc.ThresholdCfg.HighThreshold != 0.50 {
			t.Errorf("HighThreshold = %f, want 0.50", cc.ThresholdCfg.HighThreshold)
		}
		if cc.ThresholdCfg.FallbackMode != "smart" {
			t.Errorf("FallbackMode = %q, want 'smart'", cc.ThresholdCfg.FallbackMode)
		}
	})
}

// =============================================================================
// FALLBACK MESSAGE TESTS
// =============================================================================

func TestGetStaticFallbackMessage(t *testing.T) {
	service := &ChatService{}

	t.Run("uses custom message when set", func(t *testing.T) {
		cc := &chatContext{
			Bot: &models.Chatbot{
				FallbackMessages: &models.FallbackMessages{
					NoInfoFound: "Custom no info message",
				},
			},
			LangConfig: getLangConfig("tr"),
		}

		got := service.getStaticFallbackMessage(cc)
		if got != "Custom no info message" {
			t.Errorf("got %q, want custom message", got)
		}
	})

	t.Run("uses default when custom is empty", func(t *testing.T) {
		cc := &chatContext{
			Bot: &models.Chatbot{
				FallbackMessages: &models.FallbackMessages{
					NoInfoFound: "",
				},
			},
			LangConfig: getLangConfig("tr"),
		}

		got := service.getStaticFallbackMessage(cc)
		if got != "Yeterli bilgi bulamadım." {
			t.Errorf("got %q, want default Turkish message", got)
		}
	})

	t.Run("uses default when FallbackMessages is nil", func(t *testing.T) {
		cc := &chatContext{
			Bot: &models.Chatbot{
				FallbackMessages: nil,
			},
			LangConfig: getLangConfig("en"),
		}

		got := service.getStaticFallbackMessage(cc)
		if got != "I could not find enough information." {
			t.Errorf("got %q, want default English message", got)
		}
	})
}

func TestGetErrorMessage(t *testing.T) {
	service := &ChatService{}

	t.Run("uses custom error message", func(t *testing.T) {
		cc := &chatContext{
			Bot: &models.Chatbot{
				FallbackMessages: &models.FallbackMessages{
					ErrorMessage: "Custom error",
				},
			},
			LangConfig: getLangConfig("tr"),
		}

		got := service.getErrorMessage(cc)
		if got != "Custom error" {
			t.Errorf("got %q, want custom error", got)
		}
	})

	t.Run("uses default when nil", func(t *testing.T) {
		cc := &chatContext{
			Bot: &models.Chatbot{
				FallbackMessages: nil,
			},
			LangConfig: getLangConfig("tr"),
		}

		got := service.getErrorMessage(cc)
		if got != "Şu an bir hata oluştu, lütfen tekrar deneyin." {
			t.Errorf("got %q, want default Turkish error", got)
		}
	})
}

func TestGetHandoffMessage(t *testing.T) {
	service := &ChatService{}

	t.Run("uses custom handoff message", func(t *testing.T) {
		cc := &chatContext{
			Bot: &models.Chatbot{
				FallbackMessages: &models.FallbackMessages{
					HandoffMessage: "Connecting you to support...",
				},
			},
			LangConfig: getLangConfig("en"),
		}

		got := service.getHandoffMessage(cc)
		if got != "Connecting you to support..." {
			t.Errorf("got %q, want custom handoff message", got)
		}
	})
}

// =============================================================================
// APPEND USER MESSAGE WITH CONTEXT TESTS
// =============================================================================

func TestAppendUserMessageWithContext(t *testing.T) {
	service := &ChatService{}

	t.Run("adds medium tier note when conditions met", func(t *testing.T) {
		cc := &chatContext{
			Request: models.ChatRequest{Message: "What is the price?"},
			SearchResult: &rag.TieredSearchResult{
				Tier:        rag.TierMedium,
				ContextText: "Some context about products.",
			},
			ThresholdCfg: &models.ThresholdConfig{
				ShowConfidenceWarning: true,
			},
			Messages: []rag.ChatMessage{},
		}

		service.appendUserMessageWithContext(cc)

		if len(cc.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(cc.Messages))
		}

		content := *cc.Messages[0].Content
		if !contains(content, "[Note:") {
			t.Error("expected medium tier note in content")
		}
		if !contains(content, "What is the price?") {
			t.Error("expected user message in content")
		}
	})

	t.Run("no note for high tier", func(t *testing.T) {
		cc := &chatContext{
			Request: models.ChatRequest{Message: "Hello"},
			SearchResult: &rag.TieredSearchResult{
				Tier:        rag.TierHigh,
				ContextText: "Some context.",
			},
			ThresholdCfg: &models.ThresholdConfig{
				ShowConfidenceWarning: true,
			},
			Messages: []rag.ChatMessage{},
		}

		service.appendUserMessageWithContext(cc)

		content := *cc.Messages[0].Content
		if contains(content, "[Note:") {
			t.Error("should not add note for high tier")
		}
	})

	t.Run("no note when ShowConfidenceWarning is false", func(t *testing.T) {
		cc := &chatContext{
			Request: models.ChatRequest{Message: "Hello"},
			SearchResult: &rag.TieredSearchResult{
				Tier:        rag.TierMedium,
				ContextText: "Some context.",
			},
			ThresholdCfg: &models.ThresholdConfig{
				ShowConfidenceWarning: false,
			},
			Messages: []rag.ChatMessage{},
		}

		service.appendUserMessageWithContext(cc)

		content := *cc.Messages[0].Content
		if contains(content, "[Note:") {
			t.Error("should not add note when ShowConfidenceWarning is false")
		}
	})

	t.Run("no note when context is empty", func(t *testing.T) {
		cc := &chatContext{
			Request: models.ChatRequest{Message: "Hello"},
			SearchResult: &rag.TieredSearchResult{
				Tier:        rag.TierMedium,
				ContextText: "",
			},
			ThresholdCfg: &models.ThresholdConfig{
				ShowConfidenceWarning: true,
			},
			Messages: []rag.ChatMessage{},
		}

		service.appendUserMessageWithContext(cc)

		content := *cc.Messages[0].Content
		if contains(content, "[Note:") {
			t.Error("should not add note when context is empty")
		}
		// Should just be the user message
		if content != "Hello" {
			t.Errorf("expected just message, got %q", content)
		}
	})

	t.Run("adds restrictive prompt for low tier with custom actions", func(t *testing.T) {
		cc := &chatContext{
			Request: models.ChatRequest{Message: "What is the score?"},
			SearchResult: &rag.TieredSearchResult{
				Tier:        rag.TierLow,
				ContextText: "",
			},
			Actions: []*models.ChatbotAction{
				{ID: "act-1", Name: "get_score"},
			},
			Messages: []rag.ChatMessage{},
		}

		service.appendUserMessageWithContext(cc)

		content := *cc.Messages[0].Content
		if !contains(content, "[IMPORTANT: You have NO knowledge sources") {
			t.Error("expected restrictive tool-only prompt for low tier with actions")
		}
		if !contains(content, "What is the score?") {
			t.Error("expected user message in content")
		}
	})
}

// =============================================================================
// EXECUTE AGENTIC LOOP TESTS
// =============================================================================

func TestExecuteAgenticLoop_SkipConditions(t *testing.T) {
	service := &ChatService{
		Factory: rag.NewClientFactory(&config.Config{}),
	}

	t.Run("skips loop for low tier with no custom actions", func(t *testing.T) {
		cc := &chatContext{
			SearchResult: &rag.TieredSearchResult{
				Tier: rag.TierLow,
			},
			Actions: []*models.ChatbotAction{}, // No custom actions
			Tools: []rag.Tool{
				{Type: "function", Function: rag.ToolFunction{Name: "list_sources"}}, // Only built-in tool
			},
		}

		err := service.executeAgenticLoop(context.Background(), cc)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if cc.Response != "" {
			t.Error("should not set response in agent loop skip")
		}
	})

	t.Run("does not skip loop for low tier with custom actions", func(t *testing.T) {
		// This should try to perform LLM call, but we don't mock the client here
		// so it'll fail at getToolsClient, which is enough to prove it didn't skip.
		cc := &chatContext{
			SearchResult: &rag.TieredSearchResult{
				Tier: rag.TierLow,
			},
			Actions: []*models.ChatbotAction{
				{ID: "act-1", Name: "get_weather"},
			},
			Bot: &models.Chatbot{Model: "gpt-4"},
		}

		err := service.executeAgenticLoop(context.Background(), cc)

		if err == nil {
			t.Error("expected error from missing factory/client, but loop was not skipped")
		}
	})
}

// =============================================================================
// HELPER FUNCTIONS FOR TESTS
// =============================================================================

func strPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getLangConfig(code string) langconfig.LanguageConfig {
	return langconfig.Get(code)
}
