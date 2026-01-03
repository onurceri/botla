package models

import (
	"testing"
)

func TestPlanConfigValidate(t *testing.T) {
	validConfig := PlanConfig{
		Scraping: ScrapingConfig{
			DynamicEnabled:   true,
			MaxURLsPerBot:    100,
			MaxPagesPerCrawl: 50,
		},
		Files: FilesConfig{
			MaxSizeMB:      10,
			MaxFilesPerBot: 50,
			MaxFilesTotal:  500,
			TotalStorageMB: 1000,
			MaxTextLength:  100000,
		},
		Chat: ChatConfig{
			DefaultModel:     "openai/gpt-4o-mini",
			AllowedModels:    []string{"openai/gpt-4o-mini", "openai/gpt-4o"},
			MaxMonthlyTokens: 1000000,
			RAG: RAGConfig{
				TopK:             5,
				MaxContextTokens: 8000,
			},
			MaxSuggestedQuestions: 5,
			MaxManualQuestions:    5,
			MinResponseTokenLimit: 1,
			MaxResponseTokenLimit: 4096,
		},
		Refresh: RefreshConfig{
			Enabled:    true,
			MaxMonthly: 100,
		},
		Security: SecurityConfig{
			SecureEmbedEnabled: true,
		},
		Guardrails: GuardrailsConfig{
			CanCustomizeThresholds: true,
			CanUseSmartFallback:    true,
			CanUseEscalateFallback: false,
			CanManageTopics:        true,
			CanCustomizeMessages:   true,
		},
		Branding: BrandingConfig{
			CanHideBranding:   true,
			CanCustomBranding: false,
		},
		RateLimits: RateLimitsConfig{
			RequestsPerMinute: 60,
			WindowSeconds:     60,
			Endpoints: map[string]EndpointLimits{
				"/api/chat": {
					RequestsPerMinute: 30,
					WindowSeconds:     60,
				},
			},
		},
		MaxChatbots:               5,
		MaxMonthlyIngestions:      1000,
		MaxMonthlyEmbeddingTokens: 500000,
		MinReAddCooldownMinutes:   60,
	}

	t.Run("valid config", func(t *testing.T) {
		if err := validConfig.Validate(); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("max_chatbots zero", func(t *testing.T) {
		config := validConfig
		config.MaxChatbots = 0
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxChatbots = 0")
		}
	})

	t.Run("max_chatbots negative", func(t *testing.T) {
		config := validConfig
		config.MaxChatbots = -1
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxChatbots = -1")
		}
	})

	t.Run("max_monthly_tokens negative", func(t *testing.T) {
		config := validConfig
		config.Chat.MaxMonthlyTokens = -1
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxMonthlyTokens = -1")
		}
	})

	t.Run("default_model set but allowed_models empty", func(t *testing.T) {
		config := validConfig
		config.Chat.AllowedModels = []string{}
		err := config.Validate()
		if err == nil {
			t.Error("expected error when default_model is set but allowed_models is empty")
		}
	})

	t.Run("min_readd_cooldown_minutes negative", func(t *testing.T) {
		config := validConfig
		config.MinReAddCooldownMinutes = -1
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MinReAddCooldownMinutes = -1")
		}
	})

	t.Run("max_monthly_ingestions negative", func(t *testing.T) {
		config := validConfig
		config.MaxMonthlyIngestions = -1
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxMonthlyIngestions = -1")
		}
	})

	t.Run("max_monthly_embedding_tokens negative", func(t *testing.T) {
		config := validConfig
		config.MaxMonthlyEmbeddingTokens = -1
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxMonthlyEmbeddingTokens = -1")
		}
	})

	t.Run("scraping config error propagation", func(t *testing.T) {
		config := validConfig
		config.Scraping.MaxURLsPerBot = -1
		err := config.Validate()
		if err == nil {
			t.Error("expected error from scraping config")
		}
	})

	t.Run("files config error propagation", func(t *testing.T) {
		config := validConfig
		config.Files.MaxSizeMB = 0
		err := config.Validate()
		if err == nil {
			t.Error("expected error from files config")
		}
	})

	t.Run("chat config error propagation", func(t *testing.T) {
		config := validConfig
		config.Chat.MaxMonthlyTokens = -1
		err := config.Validate()
		if err == nil {
			t.Error("expected error from chat config")
		}
	})

	t.Run("refresh config error propagation", func(t *testing.T) {
		config := validConfig
		config.Refresh.MaxMonthly = -1
		err := config.Validate()
		if err == nil {
			t.Error("expected error from refresh config")
		}
	})

	t.Run("rate_limits config error propagation", func(t *testing.T) {
		config := validConfig
		config.RateLimits.RequestsPerMinute = 0
		err := config.Validate()
		if err == nil {
			t.Error("expected error from rate_limits config")
		}
	})
}

func TestScrapingConfigValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := ScrapingConfig{
			DynamicEnabled:   true,
			MaxURLsPerBot:    100,
			MaxPagesPerCrawl: 50,
		}
		if err := config.Validate(); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("max_urls_per_bot negative", func(t *testing.T) {
		config := ScrapingConfig{
			DynamicEnabled:   true,
			MaxURLsPerBot:    -1,
			MaxPagesPerCrawl: 50,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxURLsPerBot = -1")
		}
	})

	t.Run("max_pages_per_crawl negative", func(t *testing.T) {
		config := ScrapingConfig{
			DynamicEnabled:   true,
			MaxURLsPerBot:    100,
			MaxPagesPerCrawl: -1,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxPagesPerCrawl = -1")
		}
	})
}

func TestFilesConfigValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := FilesConfig{
			MaxSizeMB:      10,
			MaxFilesPerBot: 50,
			MaxFilesTotal:  500,
			TotalStorageMB: 1000,
			MaxTextLength:  100000,
		}
		if err := config.Validate(); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("max_size_mb zero", func(t *testing.T) {
		config := FilesConfig{
			MaxSizeMB:      0,
			MaxFilesPerBot: 50,
			MaxFilesTotal:  500,
			TotalStorageMB: 1000,
			MaxTextLength:  100000,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxSizeMB = 0")
		}
	})

	t.Run("max_size_mb negative", func(t *testing.T) {
		config := FilesConfig{
			MaxSizeMB:      -1,
			MaxFilesPerBot: 50,
			MaxFilesTotal:  500,
			TotalStorageMB: 1000,
			MaxTextLength:  100000,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxSizeMB = -1")
		}
	})

	t.Run("max_files_per_bot negative", func(t *testing.T) {
		config := FilesConfig{
			MaxSizeMB:      10,
			MaxFilesPerBot: -1,
			MaxFilesTotal:  500,
			TotalStorageMB: 1000,
			MaxTextLength:  100000,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxFilesPerBot = -1")
		}
	})

	t.Run("max_files_total negative", func(t *testing.T) {
		config := FilesConfig{
			MaxSizeMB:      10,
			MaxFilesPerBot: 50,
			MaxFilesTotal:  -1,
			TotalStorageMB: 1000,
			MaxTextLength:  100000,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxFilesTotal = -1")
		}
	})

	t.Run("total_storage_mb zero", func(t *testing.T) {
		config := FilesConfig{
			MaxSizeMB:      10,
			MaxFilesPerBot: 50,
			MaxFilesTotal:  500,
			TotalStorageMB: 0,
			MaxTextLength:  100000,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for TotalStorageMB = 0")
		}
	})

	t.Run("total_storage_mb negative", func(t *testing.T) {
		config := FilesConfig{
			MaxSizeMB:      10,
			MaxFilesPerBot: 50,
			MaxFilesTotal:  500,
			TotalStorageMB: -1,
			MaxTextLength:  100000,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for TotalStorageMB = -1")
		}
	})

	t.Run("max_text_length negative", func(t *testing.T) {
		config := FilesConfig{
			MaxSizeMB:      10,
			MaxFilesPerBot: 50,
			MaxFilesTotal:  500,
			TotalStorageMB: 1000,
			MaxTextLength:  -1,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxTextLength = -1")
		}
	})
}

func TestChatConfigValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := ChatConfig{
			DefaultModel:     "openai/gpt-4o-mini",
			AllowedModels:    []string{"openai/gpt-4o-mini", "openai/gpt-4o"},
			MaxMonthlyTokens: 1000000,
			RAG: RAGConfig{
				TopK:             5,
				MaxContextTokens: 8000,
			},
			MaxSuggestedQuestions: 5,
			MaxManualQuestions:    5,
			MinResponseTokenLimit: 1,
			MaxResponseTokenLimit: 4096,
		}
		if err := config.Validate(); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("no default model with empty allowed models", func(t *testing.T) {
		config := ChatConfig{
			AllowedModels:    []string{},
			MaxMonthlyTokens: 1000000,
			RAG: RAGConfig{
				TopK:             5,
				MaxContextTokens: 8000,
			},
			MaxSuggestedQuestions: 5,
			MaxManualQuestions:    5,
			MinResponseTokenLimit: 1,
			MaxResponseTokenLimit: 4096,
		}
		if err := config.Validate(); err != nil {
			t.Errorf("expected valid config with no default model and empty allowed models, got error: %v", err)
		}
	})

	t.Run("max_monthly_tokens negative", func(t *testing.T) {
		config := ChatConfig{
			MaxMonthlyTokens: -1,
			RAG: RAGConfig{
				TopK:             5,
				MaxContextTokens: 8000,
			},
			MaxSuggestedQuestions: 5,
			MaxManualQuestions:    5,
			MinResponseTokenLimit: 1,
			MaxResponseTokenLimit: 4096,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxMonthlyTokens = -1")
		}
	})

	t.Run("max_suggested_questions negative", func(t *testing.T) {
		config := ChatConfig{
			MaxMonthlyTokens:      1000000,
			MaxSuggestedQuestions: -1,
			MaxManualQuestions:    5,
			MinResponseTokenLimit: 1,
			MaxResponseTokenLimit: 4096,
			RAG: RAGConfig{
				TopK:             5,
				MaxContextTokens: 8000,
			},
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxSuggestedQuestions = -1")
		}
	})

	t.Run("max_manual_questions negative", func(t *testing.T) {
		config := ChatConfig{
			MaxMonthlyTokens:      1000000,
			MaxSuggestedQuestions: 5,
			MaxManualQuestions:    -1,
			MinResponseTokenLimit: 1,
			MaxResponseTokenLimit: 4096,
			RAG: RAGConfig{
				TopK:             5,
				MaxContextTokens: 8000,
			},
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxManualQuestions = -1")
		}
	})

	t.Run("min_response_token_limit zero", func(t *testing.T) {
		config := ChatConfig{
			MaxMonthlyTokens:      1000000,
			MaxSuggestedQuestions: 5,
			MaxManualQuestions:    5,
			MinResponseTokenLimit: 0,
			MaxResponseTokenLimit: 4096,
			RAG: RAGConfig{
				TopK:             5,
				MaxContextTokens: 8000,
			},
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MinResponseTokenLimit = 0")
		}
	})

	t.Run("min_response_token_limit negative", func(t *testing.T) {
		config := ChatConfig{
			MaxMonthlyTokens:      1000000,
			MaxSuggestedQuestions: 5,
			MaxManualQuestions:    5,
			MinResponseTokenLimit: -1,
			MaxResponseTokenLimit: 4096,
			RAG: RAGConfig{
				TopK:             5,
				MaxContextTokens: 8000,
			},
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MinResponseTokenLimit = -1")
		}
	})

	t.Run("max_response_token_limit less than min", func(t *testing.T) {
		config := ChatConfig{
			MaxMonthlyTokens:      1000000,
			MaxSuggestedQuestions: 5,
			MaxManualQuestions:    5,
			MinResponseTokenLimit: 100,
			MaxResponseTokenLimit: 50,
			RAG: RAGConfig{
				TopK:             5,
				MaxContextTokens: 8000,
			},
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error when MaxResponseTokenLimit < MinResponseTokenLimit")
		}
	})

	t.Run("rag config error propagation", func(t *testing.T) {
		config := ChatConfig{
			MaxMonthlyTokens:      1000000,
			MaxSuggestedQuestions: 5,
			MaxManualQuestions:    5,
			MinResponseTokenLimit: 1,
			MaxResponseTokenLimit: 4096,
			RAG: RAGConfig{
				TopK: 0,
			},
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error from rag config")
		}
	})
}

func TestRefreshConfigValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := RefreshConfig{
			Enabled:    true,
			MaxMonthly: 100,
		}
		if err := config.Validate(); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("max_monthly negative", func(t *testing.T) {
		config := RefreshConfig{
			Enabled:    true,
			MaxMonthly: -1,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxMonthly = -1")
		}
	})
}

func TestRateLimitsConfigValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := RateLimitsConfig{
			RequestsPerMinute: 60,
			WindowSeconds:     60,
			Endpoints: map[string]EndpointLimits{
				"/api/chat": {
					RequestsPerMinute: 30,
					WindowSeconds:     60,
				},
			},
		}
		if err := config.Validate(); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("requests_per_minute zero", func(t *testing.T) {
		config := RateLimitsConfig{
			RequestsPerMinute: 0,
			WindowSeconds:     60,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for RequestsPerMinute = 0")
		}
	})

	t.Run("requests_per_minute negative", func(t *testing.T) {
		config := RateLimitsConfig{
			RequestsPerMinute: -1,
			WindowSeconds:     60,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for RequestsPerMinute = -1")
		}
	})

	t.Run("window_seconds zero", func(t *testing.T) {
		config := RateLimitsConfig{
			RequestsPerMinute: 60,
			WindowSeconds:     0,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for WindowSeconds = 0")
		}
	})

	t.Run("window_seconds negative", func(t *testing.T) {
		config := RateLimitsConfig{
			RequestsPerMinute: 60,
			WindowSeconds:     -1,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for WindowSeconds = -1")
		}
	})

	t.Run("endpoint requests_per_minute zero", func(t *testing.T) {
		config := RateLimitsConfig{
			RequestsPerMinute: 60,
			WindowSeconds:     60,
			Endpoints: map[string]EndpointLimits{
				"/api/chat": {
					RequestsPerMinute: 0,
					WindowSeconds:     60,
				},
			},
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for endpoint RequestsPerMinute = 0")
		}
	})

	t.Run("endpoint window_seconds zero", func(t *testing.T) {
		config := RateLimitsConfig{
			RequestsPerMinute: 60,
			WindowSeconds:     60,
			Endpoints: map[string]EndpointLimits{
				"/api/chat": {
					RequestsPerMinute: 30,
					WindowSeconds:     0,
				},
			},
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for endpoint WindowSeconds = 0")
		}
	})
}

func TestRAGConfigValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := RAGConfig{
			TopK:             5,
			MaxContextTokens: 8000,
		}
		if err := config.Validate(); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("top_k zero", func(t *testing.T) {
		config := RAGConfig{
			TopK:             0,
			MaxContextTokens: 8000,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for TopK = 0")
		}
	})

	t.Run("top_k negative", func(t *testing.T) {
		config := RAGConfig{
			TopK:             -1,
			MaxContextTokens: 8000,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for TopK = -1")
		}
	})

	t.Run("max_context_tokens zero", func(t *testing.T) {
		config := RAGConfig{
			TopK:             5,
			MaxContextTokens: 0,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxContextTokens = 0")
		}
	})

	t.Run("max_context_tokens negative", func(t *testing.T) {
		config := RAGConfig{
			TopK:             5,
			MaxContextTokens: -1,
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected error for MaxContextTokens = -1")
		}
	})
}

func TestPlanConfigJSONB(t *testing.T) {
	config := PlanConfig{
		MaxChatbots: 5,
		Chat: ChatConfig{
			DefaultModel: "gpt-4",
		},
	}

	// Test Value
	value, err := config.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}

	bytes, ok := value.([]byte)
	if !ok {
		t.Fatalf("Value() did not return []byte, got %T", value)
	}

	// Test Scan
	var scanned PlanConfig
	err = scanned.Scan(bytes)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if scanned.MaxChatbots != config.MaxChatbots {
		t.Errorf("expected MaxChatbots %d, got %d", config.MaxChatbots, scanned.MaxChatbots)
	}
	if scanned.Chat.DefaultModel != config.Chat.DefaultModel {
		t.Errorf("expected DefaultModel %s, got %s", config.Chat.DefaultModel, scanned.Chat.DefaultModel)
	}

	// Test Scan with invalid type
	err = scanned.Scan(123)
	if err == nil {
		t.Error("expected error scanning non-[]byte value")
	}

	// Test Scan with invalid JSON
	err = scanned.Scan([]byte("{invalid json"))
	if err == nil {
		t.Error("expected error scanning invalid JSON")
	}
}
