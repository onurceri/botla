package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanLimits_Validate_Valid(t *testing.T) {
	limits := DefaultPlanLimits()
	err := limits.Validate()
	assert.NoError(t, err)
}

func TestPlanLimits_Validate_MaxChatbots(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"zero", 0, true},
		{"negative", -1, true},
		{"one", 1, false},
		{"many", 100, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := DefaultPlanLimits()
			limits.MaxChatbots = tt.value
			err := limits.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "max_chatbots")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanLimits_Validate_MaxMonthlyIngestions(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"negative", -1, true},
		{"zero", 0, false},
		{"positive", 100, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := DefaultPlanLimits()
			limits.MaxMonthlyIngestions = tt.value
			err := limits.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "max_monthly_ingestions")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanLimits_Validate_MaxMonthlyEmbeddingTokens(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"negative", -1, true},
		{"zero", 0, false},
		{"positive", 250000, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := DefaultPlanLimits()
			limits.MaxMonthlyEmbeddingTokens = tt.value
			err := limits.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "max_monthly_embedding_tokens")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanLimits_Validate_FilesMaxSizeMB(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"zero", 0, true},
		{"negative", -1, true},
		{"one", 1, false},
		{"large", 100, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := DefaultPlanLimits()
			limits.FilesMaxSizeMB = tt.value
			err := limits.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "files_max_size_mb")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanLimits_Validate_FilesTotalStorageMB(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"zero", 0, true},
		{"negative", -1, true},
		{"one", 1, false},
		{"large", 2000, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := DefaultPlanLimits()
			limits.FilesTotalStorageMB = tt.value
			err := limits.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "files_total_storage_mb")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanLimits_Validate_ChatRAGTopK(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"zero", 0, true},
		{"negative", -1, true},
		{"one", 1, false},
		{"ten", 10, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := DefaultPlanLimits()
			limits.ChatRAGTopK = tt.value
			err := limits.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "chat_rag_top_k")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanLimits_Validate_ChatRAGMaxContextTokens(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"zero", 0, true},
		{"negative", -1, true},
		{"one", 1, false},
		{"large", 8000, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := DefaultPlanLimits()
			limits.ChatRAGMaxContextTokens = tt.value
			err := limits.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "chat_rag_max_context_tokens")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanLimits_Validate_ResponseTokenLimits(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		max     int
		wantErr bool
		errMsg  string
	}{
		{"min_zero", 0, 4096, true, "chat_min_response_token_limit"},
		{"min_negative", -1, 4096, true, "chat_min_response_token_limit"},
		{"max_less_than_min", 100, 50, true, "chat_max_response_token_limit"},
		{"equal", 100, 100, false, ""},
		{"valid", 1, 4096, false, ""},
		{"large_valid", 1, 8192, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := DefaultPlanLimits()
			limits.ChatMinResponseTokenLimit = tt.min
			limits.ChatMaxResponseTokenLimit = tt.max
			err := limits.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanLimits_Validate_RateLimits(t *testing.T) {
	t.Run("requests_per_minute_zero", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.RateLimitsRequestsPerMinute = 0
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate_limits_requests_per_minute")
	})

	t.Run("requests_per_minute_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.RateLimitsRequestsPerMinute = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate_limits_requests_per_minute")
	})

	t.Run("window_seconds_zero", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.RateLimitsWindowSeconds = 0
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate_limits_window_seconds")
	})

	t.Run("chat_rpm_zero", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.RateLimitsChatRPM = 0
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate_limits_chat_rpm")
	})

	t.Run("chat_window_zero", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.RateLimitsChatWindow = 0
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate_limits_chat_window")
	})

	t.Run("sources_rpm_zero", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.RateLimitsSourcesRPM = 0
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate_limits_sources_rpm")
	})

	t.Run("sources_window_zero", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.RateLimitsSourcesWindow = 0
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate_limits_sources_window")
	})
}

func TestPlanLimits_Validate_MultipleErrors(t *testing.T) {
	limits := PlanLimits{
		MaxChatbots:                 0,  // invalid
		FilesMaxSizeMB:              0,  // invalid
		FilesTotalStorageMB:         0,  // invalid
		ChatRAGTopK:                 0,  // invalid
		ChatRAGMaxContextTokens:     0,  // invalid
		ChatMinResponseTokenLimit:   0,  // invalid
		ChatMaxResponseTokenLimit:   10, // would be valid if min was valid
		RateLimitsRequestsPerMinute: 0,  // invalid
		RateLimitsWindowSeconds:     0,  // invalid
		RateLimitsChatRPM:           0,  // invalid
		RateLimitsChatWindow:        0,  // invalid
		RateLimitsSourcesRPM:        0,  // invalid
		RateLimitsSourcesWindow:     0,  // invalid
	}
	err := limits.Validate()
	require.Error(t, err)

	// Should contain multiple error messages
	errStr := err.Error()
	assert.Contains(t, errStr, "max_chatbots")
	assert.Contains(t, errStr, "files_max_size_mb")
	assert.Contains(t, errStr, "files_total_storage_mb")
	assert.Contains(t, errStr, "chat_rag_top_k")
	assert.Contains(t, errStr, "chat_rag_max_context_tokens")
	assert.Contains(t, errStr, "chat_min_response_token_limit")
	assert.Contains(t, errStr, "rate_limits_requests_per_minute")
	assert.Contains(t, errStr, "rate_limits_window_seconds")
}

func TestDefaultPlanLimits(t *testing.T) {
	limits := DefaultPlanLimits()

	// Verify it passes validation
	err := limits.Validate()
	assert.NoError(t, err)

	// Verify key defaults match Free plan
	assert.Equal(t, 1, limits.MaxChatbots)
	assert.Equal(t, 50, limits.MaxMonthlyIngestions)
	assert.Equal(t, 250000, limits.MaxMonthlyEmbeddingTokens)
	assert.Equal(t, 60, limits.MinReAddCooldownMinutes)
	assert.Equal(t, "openai/gpt-4o-mini", limits.ChatDefaultModel)
	assert.Equal(t, []string{"openai/gpt-4o-mini"}, []string(limits.ChatAllowedModels))
	assert.False(t, limits.ScrapingDynamicEnabled)
	assert.False(t, limits.SecuritySecureEmbedEnabled)
	assert.False(t, limits.RefreshEnabled)
	assert.Equal(t, 100, limits.RateLimitsRequestsPerMinute)
}

func TestProPlanLimits(t *testing.T) {
	limits := ProPlanLimits()

	// Verify it passes validation
	err := limits.Validate()
	assert.NoError(t, err)

	// Verify key Pro plan values
	assert.Equal(t, 10, limits.MaxChatbots)
	assert.Equal(t, 500, limits.MaxMonthlyIngestions)
	assert.Equal(t, 2500000, limits.MaxMonthlyEmbeddingTokens)
	assert.Equal(t, "openai/gpt-4o", limits.ChatDefaultModel)
	assert.True(t, limits.ScrapingDynamicEnabled)
	assert.True(t, limits.SecuritySecureEmbedEnabled)
	assert.True(t, limits.RefreshEnabled)
	assert.Equal(t, 5, limits.RefreshMaxMonthly)
	assert.True(t, limits.BrandingCanHideBranding)
	assert.False(t, limits.BrandingCanCustomBranding)
	assert.Equal(t, 500, limits.RateLimitsRequestsPerMinute)
}

func TestUltraPlanLimits(t *testing.T) {
	limits := UltraPlanLimits()

	// Verify it passes validation
	err := limits.Validate()
	assert.NoError(t, err)

	// Verify key Ultra plan values
	assert.Equal(t, 100, limits.MaxChatbots)
	assert.Equal(t, 10000, limits.MaxMonthlyIngestions)
	assert.Equal(t, 100000000, limits.MaxMonthlyEmbeddingTokens)
	assert.Equal(t, 0, limits.MinReAddCooldownMinutes) // No cooldown for Ultra
	assert.Equal(t, "openai/gpt-4o", limits.ChatDefaultModel)
	assert.Contains(t, []string(limits.ChatAllowedModels), "openai/gpt-5")
	assert.True(t, limits.ScrapingDynamicEnabled)
	assert.True(t, limits.SecuritySecureEmbedEnabled)
	assert.True(t, limits.RefreshEnabled)
	assert.Equal(t, 100, limits.RefreshMaxMonthly)
	assert.True(t, limits.BrandingCanHideBranding)
	assert.True(t, limits.BrandingCanCustomBranding)
	assert.True(t, limits.GuardrailsCanUseEscalateFallback)
	assert.Equal(t, 2000, limits.RateLimitsRequestsPerMinute)
	assert.Equal(t, 8192, limits.ChatMaxResponseTokenLimit)
}

func TestPlanLimits_Validate_ScrapingLimits(t *testing.T) {
	t.Run("max_urls_per_bot_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.ScrapingMaxURLsPerBot = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scraping_max_urls_per_bot")
	})

	t.Run("max_urls_per_bot_zero_valid", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.ScrapingMaxURLsPerBot = 0
		err := limits.Validate()
		assert.NoError(t, err)
	})

	t.Run("max_pages_per_crawl_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.ScrapingMaxPagesPerCrawl = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scraping_max_pages_per_crawl")
	})
}

func TestPlanLimits_Validate_FilesLimits(t *testing.T) {
	t.Run("max_files_per_bot_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.FilesMaxFilesPerBot = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "files_max_files_per_bot")
	})

	t.Run("max_files_total_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.FilesMaxFilesTotal = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "files_max_files_total")
	})

	t.Run("max_text_length_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.FilesMaxTextLength = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "files_max_text_length")
	})
}

func TestPlanLimits_Validate_ChatLimits(t *testing.T) {
	t.Run("max_monthly_tokens_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.ChatMaxMonthlyTokens = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chat_max_monthly_tokens")
	})

	t.Run("max_suggested_questions_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.ChatMaxSuggestedQuestions = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chat_max_suggested_questions")
	})

	t.Run("max_manual_questions_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.ChatMaxManualQuestions = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chat_max_manual_questions")
	})
}

func TestPlanLimits_Validate_RefreshLimits(t *testing.T) {
	t.Run("max_monthly_negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.RefreshMaxMonthly = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh_max_monthly")
	})

	t.Run("max_monthly_zero_valid", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.RefreshMaxMonthly = 0
		err := limits.Validate()
		assert.NoError(t, err)
	})
}

func TestPlanLimits_Validate_MinReAddCooldown(t *testing.T) {
	t.Run("negative", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.MinReAddCooldownMinutes = -1
		err := limits.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min_readd_cooldown_minutes")
	})

	t.Run("zero_valid", func(t *testing.T) {
		limits := DefaultPlanLimits()
		limits.MinReAddCooldownMinutes = 0
		err := limits.Validate()
		assert.NoError(t, err)
	})
}
