# JSONB Migration - Remaining Work

This document details all remaining work to complete the JSONB column migration from `plans.config` to the `plan_limits` table.

## Completed Work

| Component | Status |
|-----------|--------|
| Migration files (`000052_create_plan_limits.up/down.sql`) | ✅ |
| `internal/models/plan_limits.go` | ✅ |
| `internal/models/plan_limits_test.go` (18 tests) | ✅ |
| `internal/db/plan_limits.go` | ✅ |
| `internal/db/plan_limits_test.go` (12 tests) | ✅ |
| `internal/services/plan_service.go` | ✅ |
| `internal/services/plan_service_validation_test.go` | ✅ |
| `internal/integration/fixtures/env.go` (helper + RestorePlans) | ✅ |

## Remaining Work

### Pattern for Replacement

All `jsonb_set` calls should be replaced with the new `te.UpdatePlanLimit()` helper:

```diff
# Before (old pattern)
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '100'::jsonb) WHERE code=$1`, policy.PlanFree.String())

# After (new pattern)
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "chat_max_monthly_tokens", 100)
```

### Field Name Mapping

| Old JSONB Path | New Field Name |
|----------------|----------------|
| `{max_chatbots}` | `max_chatbots` |
| `{max_monthly_ingestions}` | `max_monthly_ingestions` |
| `{chat,max_monthly_tokens}` | `chat_max_monthly_tokens` |
| `{chat,allowed_models}` | `chat_allowed_models` (use `pq.Array()`) |
| `{refresh,enabled}` | `refresh_enabled` |
| `{refresh,max_monthly}` | `refresh_max_monthly` |
| `{security,secure_embed_enabled}` | `security_secure_embed_enabled` |
| `{scraping,max_pages_per_crawl}` | `scraping_max_pages_per_crawl` |
| `{files,max_files_per_bot}` | `files_max_files_per_bot` |
| `{rate_limits,requests_per_minute}` | `rate_limits_requests_per_minute` |
| `{rate_limits,window_seconds}` | `rate_limits_window_seconds` |

---

## Files to Update

### 1. `internal/services/chat_service_validation_test.go`

**Line 34:**
```diff
-_, err = te.DB.Exec(`UPDATE plans 
-    SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '100') 
-    WHERE code=$1`, policy.PlanFree.String())
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "chat_max_monthly_tokens", 100)
```

**Line 92:**
```diff
-_, err = te.DB.Exec(`UPDATE plans 
-    SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '1000') 
-    WHERE code=$1`, policy.PlanFree.String())
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "chat_max_monthly_tokens", 1000)
```

---

### 2. `internal/api/handlers/chatbot_discovery_enforcement_test.go`

**Line 25:**
```diff
-if _, err := db.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config,'{}'::jsonb), '{scraping,max_pages_per_crawl}', '0') WHERE id=$1`, freePlanID); err != nil {
+// Need to get plan code first or use db.UpdatePlanLimitField directly
+if err := db.UpdatePlanLimitField(ctx, db, "free", "scraping_max_pages_per_crawl", 0); err != nil {
```

---

### 3. `internal/integration/plan_startup_test.go`

**Line 24:**
```diff
-_, err := db.Exec(`UPDATE plans SET config = jsonb_set(config, '{max_chatbots}', '-1'::jsonb) WHERE code = $1`, policy.PlanFree.String())
+// NOTE: This test tries to insert invalid data (-1). DB CHECK constraint will reject this.
+// The test may need to be redesigned or skipped since DB enforces validation now.
```

---

### 4. `internal/integration/ratelimit_isolation_test.go`

**Line 31:**
```diff
-_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{rate_limits}', '{"requests_per_minute": 2, "window_seconds": 60}'::jsonb) WHERE code = 'free'`)
+_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 2)
+_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)
```

---

### 5. `internal/integration/chat_config_test.go`

**Line 89 (complex - sets allowed_models and max_monthly_tokens):**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{chat}', '{"allowed_models": ["gpt-4o", "gpt-4o-mini", "anthropic:claude-3-5-sonnet-20241022", "google:gemini-1.5-flash", "openrouter:meta-llama/llama-3"], "max_monthly_tokens": 0}') WHERE code=$1`, policy.PlanFree.String())
+// Update using pq.Array for allowed_models
+import "github.com/lib/pq"
+_ = te.DB.Exec(`UPDATE plan_limits SET chat_allowed_models = $1, chat_max_monthly_tokens = 0 WHERE plan_id = (SELECT id FROM plans WHERE code = $2)`, 
+    pq.Array([]string{"gpt-4o", "gpt-4o-mini", "anthropic:claude-3-5-sonnet-20241022", "google:gemini-1.5-flash", "openrouter:meta-llama/llama-3"}),
+    policy.PlanFree.String())
```

**Line 185:** Similar pattern.

---

### 6. `internal/integration/rate_limit_test.go`

**Lines 143, 186, 225, 286, 333, 391:**
Each sets `rate_limits` object. Replace with individual field updates:

```diff
-_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{rate_limits}', '{"requests_per_minute": 4, "window_seconds": 60}'::jsonb) WHERE code = 'free'`)
+_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 4)
+_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)
```

---

### 7. `internal/integration/temperature_model_test.go`

**Line 372:**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{chat,allowed_models}', '["gpt-4o", "gpt-4o-mini", "anthropic:claude-3-5-sonnet", "google:gemini-1.5-flash", "openrouter:meta-llama/llama-3"]') WHERE code=$1`, policy.PlanFree.String())
+_ = te.DB.Exec(`UPDATE plan_limits SET chat_allowed_models = $1 WHERE plan_id = (SELECT id FROM plans WHERE code = $2)`,
+    pq.Array([]string{"gpt-4o", "gpt-4o-mini", "anthropic:claude-3-5-sonnet", "google:gemini-1.5-flash", "openrouter:meta-llama/llama-3"}),
+    policy.PlanFree.String())
```

**Line 472:** Similar pattern.

---

### 8. `internal/integration/plan_limits_free_test.go`

**Line 411 (refresh config):**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{refresh}', '{"enabled": false, "max_monthly": 0}'::jsonb, true) WHERE code = '` + policy.PlanFree.String() + `'`)
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "refresh_enabled", false)
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "refresh_max_monthly", 0)
```

**Line 512 (security config):**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": false}'::jsonb, true) WHERE code = '` + policy.PlanFree.String() + `'`)
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "security_secure_embed_enabled", false)
```

**Line 574:** Similar pattern.

---

### 9. `internal/integration/dedup_test.go`

**Line 29:**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{files,max_files_per_bot}', '10'::jsonb) WHERE code = 'free'`)
+_ = te.UpdatePlanLimit("free", "files_max_files_per_bot", 10)
```

---

### 10. `internal/integration/public_secure_embed_test.go`

**Line 27:**
```diff
-_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code=$1`, policy.PlanFree.String())
+err = te.UpdatePlanLimit(policy.PlanFree.String(), "security_secure_embed_enabled", true)
```

---

### 11. `internal/integration/chatbot_secure_embed_test.go`

**Line 24:**
```diff
-_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code=$1`, policy.PlanFree.String())
+err = te.UpdatePlanLimit(policy.PlanFree.String(), "security_secure_embed_enabled", true)
```

---

### 12. `internal/integration/ratelimit_headers_test.go`

**Line 32:**
```diff
-_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{rate_limits}', '{"requests_per_minute": 3, "window_seconds": 60}'::jsonb) WHERE code = 'free'`)
+_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 3)
+_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)
```

---

### 13. `internal/integration/source_refresh_test.go`

**Lines 40, 144, 189, 233:**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{refresh}', '{"enabled": true, "max_monthly": 5}'::jsonb) WHERE code = 'free'`)
+_ = te.UpdatePlanLimit("free", "refresh_enabled", true)
+_ = te.UpdatePlanLimit("free", "refresh_max_monthly", 5)
```

---

### 14. `internal/integration/secure_embed_comprehensive_test.go`

**Lines 25, 103, 174, 234, 294, 391:**
```diff
-_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(COALESCE(config, '{}'::jsonb), '{security}', '{"secure_embed_enabled": true}'::jsonb, true) WHERE code=$1`, policy.PlanFree.String())
+err = te.UpdatePlanLimit(policy.PlanFree.String(), "security_secure_embed_enabled", true)
```

---

### 15. `internal/integration/source_refresh_hash_test.go`

**Line 44:**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{refresh}', '{"enabled": true, "max_monthly": 10}'::jsonb) WHERE code = 'free'`)
+_ = te.UpdatePlanLimit("free", "refresh_enabled", true)
+_ = te.UpdatePlanLimit("free", "refresh_max_monthly", 10)
```

---

### 16. `internal/integration/turkish_test.go`

**Line 172:**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '100'::jsonb) WHERE code=$1`, policy.PlanFree.String())
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "chat_max_monthly_tokens", 100)
```

---

### 17. `internal/integration/ratelimit_test.go`

**Line 31:**
```diff
-_, err = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{rate_limits}', '{"requests_per_minute": 4, "window_seconds": 60}'::jsonb) WHERE code = 'free'`)
+_ = te.UpdatePlanLimit("free", "rate_limits_requests_per_minute", 4)
+_ = te.UpdatePlanLimit("free", "rate_limits_window_seconds", 60)
```

---

### 18. `internal/integration/quota_enforcement_test.go`

**Line 49:**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '100'::jsonb) WHERE code=$1`, policy.PlanFree.String())
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "chat_max_monthly_tokens", 100)
```

**Lines 123-124 (refresh):**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{refresh,enabled}', 'true'::jsonb) WHERE code=$1`, policy.PlanFree.String())
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{refresh,max_monthly}', '1'::jsonb) WHERE code=$1`, policy.PlanFree.String())
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "refresh_enabled", true)
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "refresh_max_monthly", 1)
```

**Line 183:**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{max_monthly_ingestions}', '1'::jsonb) WHERE code=$1`, policy.PlanFree.String())
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "max_monthly_ingestions", 1)
```

**Line 256:**
```diff
-_, _ = te.DB.Exec(`UPDATE plans SET config = jsonb_set(config, '{chat,max_monthly_tokens}', '1000'::jsonb) WHERE code=$1`, policy.PlanFree.String())
+_ = te.UpdatePlanLimit(policy.PlanFree.String(), "chat_max_monthly_tokens", 1000)
```

---

## Special Cases

### Tests that try to insert invalid data

**`plan_startup_test.go` Line 24** tries to set `max_chatbots = -1`. The database CHECK constraint `chk_max_chatbots` will reject this. The test needs to be:
1. **Redesigned** to test validation at a different layer, OR
2. **Skipped** with a comment explaining DB-level validation

### Tests that update `chat_allowed_models` (array field)

The `chat_allowed_models` field is a `TEXT[]` PostgreSQL array. To update it:

```go
import "github.com/lib/pq"

_, _ = te.DB.Exec(`UPDATE plan_limits SET chat_allowed_models = $1 WHERE plan_id = (SELECT id FROM plans WHERE code = $2)`,
    pq.Array([]string{"model1", "model2"}),
    planCode)
```

Or add a new helper method:
```go
func (te *TestEnv) UpdatePlanLimitArray(planCode, field string, values []string) error {
    return te.DB.Exec(`UPDATE plan_limits SET `+field+` = $1 WHERE plan_id = (SELECT id FROM plans WHERE code = $2)`,
        pq.Array(values), planCode)
}
```

---

## Cleanup After All Files Updated

After all integration tests are updated:

1. **Remove deprecated code from `internal/models/plan.go`:**
   - Remove `PlanConfig` struct and all nested config structs
   - Remove `Value()` and `Scan()` methods
   - Change `Config PlanConfig` field to just use `Limits *PlanLimits`

2. **Run final verification:**
   ```bash
   make test-all
   make lint
   grep -r "jsonb_set" internal/ --include="*.go" | wc -l  # Should be 0
   ```

3. **Update frontend types:**
   - `frontend/src/hooks/queries/usePlans.ts`

---

## Summary

| Category | Count |
|----------|-------|
| Integration test files to update | 18 |
| Total `jsonb_set` occurrences | ~40 |
| Model cleanup items | 3 |
| Frontend files | 1 |
