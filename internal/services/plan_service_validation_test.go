package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
)

func TestPlanService_ValidateAllPlans_EmptyDatabase(t *testing.T) {
	db := testdb.OpenParallelTestDB(t)
	svc := NewPlanService(db, nil)
	ctx := context.Background()

	plans, err := svc.GetAllPlans(ctx)
	if err != nil {
		t.Skipf("skipping: could not fetch plans: %v", err)
	}

	if len(plans) == 0 {
		err = svc.ValidateAllPlans(ctx)
		if err != nil {
			t.Errorf("expected no error for empty database, got %v", err)
		}
	}
}

func TestPlanService_ValidateAllPlans_AllPlansValid(t *testing.T) {
	db := testdb.OpenParallelTestDB(t)
	svc := NewPlanService(db, nil)
	ctx := context.Background()

	validPlanConfig := models.PlanConfig{
		MaxChatbots:               1,
		MaxMonthlyIngestions:      1000,
		MaxMonthlyEmbeddingTokens: 1000000,
		MinReAddCooldownMinutes:   60,
		Scraping: models.ScrapingConfig{
			MaxURLsPerBot:    100,
			MaxPagesPerCrawl: 10,
		},
		Files: models.FilesConfig{
			MaxSizeMB:      10,
			MaxFilesPerBot: 50,
			MaxFilesTotal:  500,
			TotalStorageMB: 1000,
			MaxTextLength:  100000,
		},
		Chat: models.ChatConfig{
			DefaultModel:     "openai/gpt-4o-mini",
			AllowedModels:    []string{"openai/gpt-4o-mini"},
			MaxMonthlyTokens: 1000000,
			RAG: models.RAGConfig{
				TopK:             5,
				MaxContextTokens: 4096,
			},
			MaxSuggestedQuestions: 3,
			MaxManualQuestions:    3,
			MinResponseTokenLimit: 1,
			MaxResponseTokenLimit: 4096,
		},
		Refresh: models.RefreshConfig{
			Enabled:    true,
			MaxMonthly: 100,
		},
		RateLimits: models.RateLimitsConfig{
			RequestsPerMinute: 60,
			WindowSeconds:     60,
		},
	}

	configBytes, _ := json.Marshal(validPlanConfig)

	planID := "11111111-1111-1111-1111-111111111111"
	_, err := db.ExecContext(ctx, `
		INSERT INTO plans (id, code, status, billing_cycle, price, currency, trial_days, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, planID, "valid_test_plan", "active", "monthly", 29.99, "USD", 14, configBytes, time.Now())
	if err != nil {
		t.Fatalf("failed to insert test plan: %v", err)
	}

	plans, err := svc.GetAllPlans(ctx)
	if err != nil {
		t.Fatalf("failed to fetch plans: %v", err)
	}

	foundValidPlan := false
	for _, p := range plans {
		if p.Code == "valid_test_plan" {
			foundValidPlan = true
			if err := p.Config.Validate(); err != nil {
				t.Errorf("valid_test_plan should be valid, got error: %v", err)
			}
			break
		}
	}
	if !foundValidPlan {
		t.Error("valid_test_plan not found in fetched plans")
	}
}

func TestPlanService_ValidateAllPlans_InvalidPlanConfig(t *testing.T) {
	db := testdb.OpenParallelTestDB(t)
	svc := NewPlanService(db, nil)
	ctx := context.Background()

	invalidPlanConfig := models.PlanConfig{
		MaxChatbots:          0,
		MaxMonthlyIngestions: -1,
		Scraping: models.ScrapingConfig{
			MaxURLsPerBot: -1,
		},
		Files: models.FilesConfig{
			MaxSizeMB: -1,
		},
		Chat: models.ChatConfig{
			MaxMonthlyTokens: -1,
			RAG: models.RAGConfig{
				TopK: -1,
			},
		},
		RateLimits: models.RateLimitsConfig{
			RequestsPerMinute: -1,
		},
	}

	configBytes, _ := json.Marshal(invalidPlanConfig)

	planID := "22222222-2222-2222-2222-222222222222"
	_, err := db.ExecContext(ctx, `
		INSERT INTO plans (id, code, status, billing_cycle, price, currency, trial_days, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, planID, "invalid_test_plan", "active", "monthly", 9.99, "USD", 0, configBytes, time.Now())
	if err != nil {
		t.Fatalf("failed to insert invalid test plan: %v", err)
	}

	err = svc.ValidateAllPlans(ctx)
	if err == nil {
		t.Error("expected error for invalid plan config, got nil")
	}

	if err != nil {
		errStr := err.Error()
		if !strings.Contains(errStr, "plan \"invalid_test_plan\"") {
			t.Errorf("expected error to mention plan code, got %s", errStr)
		}
	}
}

func TestPlanService_ValidateAllPlans_MultiplePlansWithMixedValidity(t *testing.T) {
	db := testdb.OpenParallelTestDB(t)
	svc := NewPlanService(db, nil)
	ctx := context.Background()

	validConfig := models.PlanConfig{
		MaxChatbots:               1,
		MaxMonthlyIngestions:      100,
		MaxMonthlyEmbeddingTokens: 10000,
		MinReAddCooldownMinutes:   30,
		Scraping: models.ScrapingConfig{
			MaxURLsPerBot:    10,
			MaxPagesPerCrawl: 5,
		},
		Files: models.FilesConfig{
			MaxSizeMB:      5,
			MaxFilesPerBot: 10,
			MaxFilesTotal:  100,
			TotalStorageMB: 100,
			MaxTextLength:  10000,
		},
		Chat: models.ChatConfig{
			DefaultModel:          "openai/gpt-4o-mini",
			AllowedModels:         []string{"openai/gpt-4o-mini"},
			MaxMonthlyTokens:      100000,
			RAG:                   models.RAGConfig{TopK: 3, MaxContextTokens: 2048},
			MaxSuggestedQuestions: 3,
			MaxManualQuestions:    3,
			MinResponseTokenLimit: 1,
			MaxResponseTokenLimit: 2048,
		},
		Refresh: models.RefreshConfig{
			Enabled:    true,
			MaxMonthly: 10,
		},
		RateLimits: models.RateLimitsConfig{
			RequestsPerMinute: 30,
			WindowSeconds:     60,
		},
	}

	validConfigBytes, _ := json.Marshal(validConfig)
	invalidConfigBytes := []byte(`{"max_chatbots": 0, "max_monthly_ingestions": -1}`)

	planID1 := "33333333-3333-3333-3333-333333333333"
	_, err := db.ExecContext(ctx, `
		INSERT INTO plans (id, code, status, billing_cycle, price, currency, trial_days, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, planID1, "valid_plan_1", "active", "monthly", 19.99, "USD", 7, validConfigBytes, time.Now())
	if err != nil {
		t.Fatalf("failed to insert valid plan 1: %v", err)
	}

	planID2 := "44444444-4444-4444-4444-444444444444"
	_, err = db.ExecContext(ctx, `
		INSERT INTO plans (id, code, status, billing_cycle, price, currency, trial_days, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, planID2, "invalid_plan_multi", "active", "monthly", 29.99, "USD", 14, invalidConfigBytes, time.Now())
	if err != nil {
		t.Fatalf("failed to insert invalid plan: %v", err)
	}

	planID3 := "55555555-5555-5555-5555-555555555555"
	_, err = db.ExecContext(ctx, `
		INSERT INTO plans (id, code, status, billing_cycle, price, currency, trial_days, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, planID3, "valid_plan_2", "active", "monthly", 99.99, "USD", 0, validConfigBytes, time.Now())
	if err != nil {
		t.Fatalf("failed to insert valid plan 2: %v", err)
	}

	err = svc.ValidateAllPlans(ctx)
	if err == nil {
		t.Error("expected error for mixed valid/invalid plans, got nil")
	}

	if err != nil {
		errStr := err.Error()
		if !strings.Contains(errStr, "invalid_plan_multi") {
			t.Errorf("expected error to mention invalid plan, got %s", errStr)
		}
	}
}

func TestPlanService_ValidateAllPlans_NullConfig(t *testing.T) {
	db := testdb.OpenParallelTestDB(t)
	svc := NewPlanService(db, nil)
	ctx := context.Background()

	planID := "66666666-6666-6666-6666-666666666666"
	_, err := db.ExecContext(ctx, `
		INSERT INTO plans (id, code, status, billing_cycle, price, currency, trial_days, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, planID, "null_config_plan", "active", "monthly", 9.99, "USD", 0, nil, time.Now())
	if err != nil {
		t.Fatalf("failed to insert plan with null config: %v", err)
	}

	err = svc.ValidateAllPlans(ctx)
	if err == nil {
		t.Error("expected error for null config, got nil")
	}

	if err != nil {
		errStr := err.Error()
		if !strings.Contains(errStr, "null_config_plan") && !strings.Contains(errStr, "Scan error") {
			t.Errorf("expected error to mention plan with null config or scan error, got %s", errStr)
		}
	}
}

func TestPlanService_ValidateAllPlans_InvalidValuesConfig(t *testing.T) {
	db := testdb.OpenParallelTestDB(t)
	svc := NewPlanService(db, nil)
	ctx := context.Background()

	invalidValuesJSON := []byte(`{"max_chatbots": 0, "max_monthly_ingestions": -1}`)

	planID := "77777777-7777-7777-7777-777777777777"
	_, err := db.ExecContext(ctx, `
		INSERT INTO plans (id, code, status, billing_cycle, price, currency, trial_days, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, planID, "invalid_values_plan", "active", "monthly", 9.99, "USD", 0, invalidValuesJSON, time.Now())
	if err != nil {
		t.Fatalf("failed to insert plan with invalid config values: %v", err)
	}

	err = svc.ValidateAllPlans(ctx)
	if err == nil {
		t.Error("expected error for invalid config values, got nil")
	}

	if err != nil {
		errStr := err.Error()
		if !strings.Contains(errStr, "invalid_values_plan") {
			t.Errorf("expected error to mention plan with invalid values, got %s", errStr)
		}
	}
}

func TestPlanService_ValidateAllPlans_UnmarshalError(t *testing.T) {
	db := testdb.OpenParallelTestDB(t)
	svc := NewPlanService(db, nil)
	ctx := context.Background()

	// "[]" is valid JSON but cannot be unmarshaled into PlanConfig struct (which is an object)
	invalidJsonForStruct := []byte(`[]`) 

	planID := "88888888-8888-8888-8888-888888888888"
	_, err := db.ExecContext(ctx, `
		INSERT INTO plans (id, code, status, billing_cycle, price, currency, trial_days, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, planID, "unmarshal_error_plan", "active", "monthly", 9.99, "USD", 0, invalidJsonForStruct, time.Now())
	if err != nil {
		t.Fatalf("failed to insert plan with incompatible JSON: %v", err)
	}

	err = svc.ValidateAllPlans(ctx)
	if err == nil {
		t.Error("expected error for unmarshal failure, got nil")
	}

	if err != nil {
		errStr := err.Error()
		if !strings.Contains(errStr, "unmarshal_error_plan") && !strings.Contains(errStr, "unmarshal") {
			t.Errorf("expected error to mention plan with unmarshal error, got %s", errStr)
		}
	}
}

func TestPlanService_ValidateAllPlans_DatabaseError(t *testing.T) {
	db := testdb.OpenParallelTestDB(t)
	svc := NewPlanService(db, nil)
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	err := svc.ValidateAllPlans(ctx)
	if err == nil {
		t.Error("expected error for database failure, got nil")
	}
}
