# JSONB Migration - Complete Plan for Full Migration

This document outlines the complete plan to migrate all remaining code from using the deprecated `plans.config` JSONB column to the new normalized `plan_limits` table.

## Current State

### ✅ Already Completed (Phase 1)
- Migration files (`000052_create_plan_limits.up/down.sql`)
- `internal/models/plan_limits.go` with PlanLimits model
- `internal/db/plan_limits.go` with DB helpers (GetPlanLimitsByPlanID, GetPlanLimitsByCode, UpdatePlanLimitField, GetPlanWithLimits, GetAllPlansWithLimits)
- `internal/services/plan_service.go` with plan validation
- Integration test helper `fixtures.RestorePlans()` and `te.UpdatePlanLimit()`
- 18 integration test files updated to use new helper methods

### ❌ Still Needing Migration (Phase 2)
- Database query layer still selects `plans.config`
- Business logic layer still accesses `plan.Config` fields
- API handler layer still returns `plan.Config` in responses
- Integration tests with `jsonb_set` patterns (completed in Phase 1)

---

## Phase 2: Core Code Migration

### 1. Database Layer (`internal/db/plan.go`)

**File:** `internal/db/plan.go`

**Current Code (Line 14):**
```go
SELECT p.id, p.code, p.status, p.billing_cycle, p.price, p.currency, p.trial_days, p.config, p.created_at, p.updated_at
FROM plans p
JOIN users u ON u.plan_id = p.id
WHERE u.id = $1 AND u.deleted_at IS NULL AND p.deleted_at IS NULL
```

**Changes Needed:**
1. Remove `p.config` from SELECT statement
2. Add JOIN to `plan_limits` table
3. Scan limits into `plan.Limits` instead of config

**New Implementation:**
```go
SELECT p.id, p.code, p.status, p.billing_cycle, p.price, p.currency, p.trial_days, 
       p.created_at, p.updated_at,
       pl.max_chatbots, pl.max_monthly_ingestions, pl.max_monthly_embedding_tokens,
       pl.min_readd_cooldown_minutes, pl.scraping_dynamic_enabled, pl.scraping_max_urls_per_bot,
       pl.scraping_max_pages_per_crawl, pl.files_max_size_mb, pl.files_max_files_per_bot,
       pl.files_max_files_total, pl.files_total_storage_mb, pl.files_max_text_length,
       pl.chat_default_model, pl.chat_allowed_models, pl.chat_max_monthly_tokens,
       pl.chat_rag_top_k, pl.chat_rag_max_context_tokens, pl.chat_max_suggested_questions,
       pl.chat_max_manual_questions, pl.chat_min_response_token_limit, pl.chat_max_response_token_limit,
       pl.refresh_enabled, pl.refresh_max_monthly, pl.security_secure_embed_enabled,
       pl.guardrails_can_customize_thresholds, pl.guardrails_can_use_smart_fallback,
       pl.guardrails_can_use_escalate_fallback, pl.guardrails_can_manage_topics,
       pl.guardrails_can_customize_messages, pl.branding_can_hide_branding,
       pl.branding_can_custom_branding, pl.rate_limits_requests_per_minute,
       pl.rate_limits_window_seconds, pl.rate_limits_chat_rpm, pl.rate_limits_chat_window,
       pl.rate_limits_sources_rpm, pl.rate_limits_sources_window
FROM plans p
JOIN users u ON u.plan_id = p.id
LEFT JOIN plan_limits pl ON pl.plan_id = p.id
WHERE u.id = $1 AND u.deleted_at IS NULL AND p.deleted_at IS NULL
```

---

### 2. API Handler Layer

#### 2.1 `internal/api/handlers/plan.go`

**Current Issues:**
- Line 150: SELECT includes `p.config`
- Lines 106-119: Maps `plan.Config` fields to response

**Changes:**
1. Remove config from SELECT
2. Change mapping from `plan.Config.X` to `plan.Limits.X`

**Field Mapping:**
| Old (`plan.Config`) | New (`plan.Limits`) |
|--------------------|---------------------|
| `Config.MaxChatbots` | `Limits.MaxChatbots` |
| `Config.MaxMonthlyIngestions` | `Limits.MaxMonthlyIngestions` |
| `Config.Scraping` | Use individual limits |
| `Config.Files` | Use individual limits |
| `Config.Chat` | Use individual limits |
| `Config.Refresh` | Use `Limits.RefreshEnabled`, `Limits.RefreshMaxMonthly` |
| `Config.Security` | Use `Limits.SecuritySecureEmbedEnabled` |
| `Config.Guardrails` | Use individual guardrail fields |
| `Config.Branding` | Use `Limits.BrandingCanHideBranding`, `Limits.BrandingCanCustomBranding` |
| `Config.RateLimits` | Use rate limit fields |

#### 2.2 `internal/api/handlers/plans_handler.go`

**Current Issues:**
- Lines 86-98: Maps `plan.Config` to API response

**Changes:**
- Change all field accesses to use `plan.Limits`

---

### 3. Business Logic Layer

#### 3.1 Middleware (`pkg/middleware/ratelimit.go`)

**Current (Line 63):**
```go
rateLimitsCfg := plan.Config.RateLimits
```

**Change to:**
```go
rateLimitsCfg := ratelimit.Config{
    RequestsPerMinute: plan.Limits.RateLimitsRequestsPerMinute,
    WindowSeconds:     plan.Limits.RateLimitsWindowSeconds,
}
```

#### 3.2 Validation (`internal/validation/chatbot_validator.go`)

**Current Pattern (14 occurrences):**
```go
minLimit := plan.Config.Chat.MinResponseTokenLimit
maxLimit := plan.Config.Chat.MaxResponseTokenLimit
// etc.
```

**Changes:**
| Old | New |
|-----|-----|
| `plan.Config.Chat.MinResponseTokenLimit` | `plan.Limits.ChatMinResponseTokenLimit` |
| `plan.Config.Chat.MaxResponseTokenLimit` | `plan.Limits.ChatMaxResponseTokenLimit` |
| `plan.Config.Chat.MaxManualQuestions` | `plan.Limits.ChatMaxManualQuestions` |
| `plan.Config.Chat.AllowedModels` | `plan.Limits.ChatAllowedModels` |
| `plan.Config.Branding.CanHideBranding` | `plan.Limits.BrandingCanHideBranding` |
| `plan.Config.Branding.CanCustomBranding` | `plan.Limits.BrandingCanCustomBranding` |
| `plan.Config.Refresh.Enabled` | `plan.Limits.RefreshEnabled` |
| `plan.Config.Refresh.MaxMonthly` | `plan.Limits.RefreshMaxMonthly` |
| `plan.Config.Scraping.MaxPagesPerCrawl` | `plan.Limits.ScrapingMaxPagesPerCrawl` |
| `plan.Config.Security.SecureEmbedEnabled` | `plan.Limits.SecuritySecureEmbedEnabled` |
| `plan.Config.Guardrails` | Use individual guardrail fields |

#### 3.3 Services (`internal/services/`)

**Files to update:**
- `chat_service.go` (lines 84-151): Change `plan.Config.X` → `plan.Limits.X`
- `chat_helpers.go` (line 119): Change `plan.Config.Guardrails.CanUseEscalateFallback`
- `refresh_scheduler.go` (line 202): Change `plan.Config.Refresh.MaxMonthly`

#### 3.4 Processing (`internal/processing/`)

**Files to update:**
- `url_processor.go` (lines 138-447): 10+ accesses to `plan.Config`
- `pdf_processor.go` (lines 108-109): `plan.Config.Chat.MaxSuggestedQuestions`
- `text_processor.go` (lines 89-90): `plan.Config.Chat.MaxSuggestedQuestions`

#### 3.5 Direct SQL Access (`internal/processing/suggestions.go`)

**Current (Line 203):**
```sql
SELECT COALESCE((p.config->'chat'->>'max_suggested_questions')::int, $2)
```

**Change to:**
```sql
SELECT COALESCE(pl.chat_max_suggested_questions, $2)
FROM plans p
LEFT JOIN plan_limits pl ON pl.plan_id = p.id
```

---

## Phase 3: Model Cleanup

After all usages are migrated, clean up the deprecated code:

### `internal/models/plan.go`

**Remove:**
1. `PlanConfig` struct (lines 26-39)
2. `ScrapingConfig`, `FilesConfig`, `ChatConfig`, `RefreshConfig`, `SecurityConfig`, `GuardrailsConfig`, `BrandingConfig`, `RateLimitsConfig` structs
3. `Value()` method (lines 107-114)
4. `Scan()` method (lines 116-126)
5. `Validate()` methods (lines 128-246)

**Update:**
```go
type Plan struct {
    ID           string      `json:"id"`
    Code         string      `json:"code"`
    Status       string      `json:"status"`
    BillingCycle string      `json:"billing_cycle"`
    Price        float64     `json:"price"`
    Currency     string      `json:"currency"`
    TrialDays    int         `json:"trial_days"`
    Limits       *PlanLimits `json:"limits"`     // Changed from deprecated Config
    CreatedAt    time.Time   `json:"created_at"`
    UpdatedAt    *time.Time  `json:"updated_at"`
}
```

---

## Phase 4: Frontend Updates

**File:** `frontend/src/hooks/queries/usePlans.ts`

Update type definitions and API response handling to use new `limits` field instead of `config`.

---

## Verification Commands

```bash
# 1. Check for remaining jsonb_set calls (should be 0)
grep -r "jsonb_set" internal/ --include="*.go" | wc -l

# 2. Check for remaining plan.Config accesses
grep -r "plan\.Config\." internal/ --include="*.go" | wc -l
grep -r "p\.Config\." internal/ --include="*.go" | wc -l

# 3. Check for PlanConfig struct usage
grep -r "PlanConfig" internal/ --include="*.go" | grep -v "_test.go" | wc -l

# 4. Run tests
make test-all
```

---

## File Checklist

### Phase 2: Core Code Migration

| File | Status | Changes Needed |
|------|--------|----------------|
| `internal/db/plan.go` | ❌ Not started | Remove p.config from SELECT, JOIN plan_limits |
| `internal/api/handlers/plan.go` | ❌ Not started | Remove config from SELECT, use plan.Limits |
| `internal/api/handlers/plans_handler.go` | ❌ Not started | Use plan.Limits instead of plan.Config |
| `pkg/middleware/ratelimit.go` | ❌ Not started | Use plan.Limits.RateLimits* |
| `internal/validation/chatbot_validator.go` | ❌ Not started | 14 field accesses to update |
| `internal/services/chat_service.go` | ❌ Not started | 7 field accesses to update |
| `internal/services/chat_helpers.go` | ❌ Not started | 1 field access to update |
| `internal/services/refresh_scheduler.go` | ❌ Not started | 1 field access to update |
| `internal/processing/url_processor.go` | ❌ Not started | 10+ field accesses to update |
| `internal/processing/pdf_processor.go` | ❌ Not started | 2 field accesses to update |
| `internal/processing/text_processor.go` | ❌ Not started | 2 field accesses to update |
| `internal/processing/suggestions.go` | ❌ Not started | Change SQL query |
| `internal/api/handlers/source_create.go` | ❌ Not started | 4 field accesses to update |
| `internal/api/handlers/source_refresh.go` | ❌ Not started | 2 field accesses to update |
| `internal/api/handlers/source_utils.go` | ❌ Not started | 3 field accesses to update |
| `internal/api/handlers/chatbot_list.go` | ❌ Not started | 2 field accesses to update |
| `internal/api/handlers/source_bulk.go` | ❌ Not started | 1 field access to update |
| `internal/api/handlers/public.go` | ❌ Not started | 4 field accesses to update |

### Phase 3: Model Cleanup

| File | Status | Changes Needed |
|------|--------|----------------|
| `internal/models/plan.go` | ❌ Not started | Remove PlanConfig and all nested structs |

### Phase 4: Frontend

| File | Status | Changes Needed |
|------|--------|----------------|
| `frontend/src/hooks/queries/usePlans.ts` | ❌ Not started | Update types and response handling |

---

## Estimated Effort

| Phase | Files | Estimated Time |
|-------|-------|----------------|
| Phase 2: Core Code Migration | 17 | 4-6 hours |
| Phase 3: Model Cleanup | 1 | 30 minutes |
| Phase 4: Frontend | 1 | 15 minutes |
| **Total** | **19** | **5-7 hours** |

---

## Rollback Plan

If issues arise during migration:

1. Keep the migration files (000052) as they only INSERT if not exists
2. The `plan.Config` field can remain in the model temporarily
3. Tests can be run incrementally after each file change
4. Use `git stash` to save progress if needed

