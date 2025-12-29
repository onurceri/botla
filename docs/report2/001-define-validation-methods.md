# TASK-001 — Define Strict Validation for Plan Config Models

Goal:
Implement a comprehensive `Validate()` method for `PlanConfig` and all nested configuration structs to ensure type safety and logical correctness of plan limits and feature flags.

Scope:
- Modify `internal/models/plan.go` to add `Validate()` methods.
- Define validation rules for `ScrapingConfig`, `FilesConfig`, `ChatConfig`, `RefreshConfig`, `SecurityConfig`, `GuardrailsConfig`, `BrandingConfig`, and `RateLimitsConfig`.
- Include unit tests for all validation logic.

Checklist:
- [x] Identify happy paths for each configuration sub-struct (e.g., positive limits, valid enums).
- [x] Identify edge cases (negative numbers, zero values where positive is required, empty lists).
- [x] Create `internal/models/plan_validation_test.go`.
- [x] Write failing unit tests for `PlanConfig.Validate()` covering all sub-configs.
- [x] Implement `Validate()` methods in `internal/models/plan.go` to satisfy tests.
- [x] Ensure all error messages are descriptive (e.g., "scraping.max_urls_per_bot must be >= 0").
- [x] Refactor for clarity.
- [x] Run `go test ./internal/models/...` and ensure all pass.
- [x] Run `golangci-lint run ./internal/models/`.

Edge Cases:
- `MaxChatbots <= 0`: Plans should allow at least one chatbot or explicit enterprise-level definitions.
- `MaxMonthlyTokens < 0`: Usage limits cannot be negative.
- `AllowedModels`: Should not be empty if a default model is specified.
- `MinReAddCooldownMinutes < 0`: Cooldowns must be non-negative.
- `RateLimits`: Request counts and windows must be positive.

Files Likely to Change:
- `internal/models/plan.go`
- `internal/models/plan_validation_test.go` (New)
