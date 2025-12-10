# Plan: Fix Secure Embed Enforcement

## Problem
The "Secure Embed" feature (whitelisting domains and using an embed secret) is intended for paid plans (Pro and Ultra). Currently, the Frontend hides this option for Free plan users, but the Backend API (`UpdateChatbot`) does not enforce this restriction. A malicious user could manually send a request to enable secure embed on a Free plan.

## Analysis
- **File:** `internal/api/handlers/chatbot_item.go`
- **Function:** `updateChatbot` / `applyChatbotUpdates`
- **Current State:** The handler blindly accepts `SecureEmbedEnabled`, `AllowedDomains`, and `EmbedSecret` from the request body.
- **Requirement:** Check the user's plan configuration before applying these updates.

## Proposed Changes

### 1. Update `updateChatbot` Handler
- Retrieve the user's plan using `db.GetPlanByUserID`.
- Check if `SecureEmbedEnabled` is being set to `true`.
- If the plan is "free" (or specifically checks a config flag if available, though currently it seems hardcoded to plan tiers in FE logic), reject the request or ignore the field.
- **Better Approach:** Add a `secure_embed` boolean to `PlanConfig` in the database (similar to `ocr_enabled`) to make it data-driven, rather than hardcoding "free" check.

### 2. Database Migration
- Add `secure_embed_enabled` to `plans.config` JSON structure.
- Update `000003_update_plan_configs.up.sql` (or create a new migration) to set this to `false` for Free and `true` for Pro/Ultra.

### 3. Backend Implementation
- In `internal/models/plan.go`, update `PlanConfig` struct to include `SecureEmbedEnabled` (likely under `ChatConfig` or a new `SecurityConfig` section, or just `Branding`?). Let's put it under a new `Security` key or existing `Chat` key.
- In `internal/api/handlers/chatbot_item.go`, inside `updateChatbot`:
    ```go
    if req.SecureEmbedEnabled != nil && *req.SecureEmbedEnabled {
        // Check plan config
        if !plan.Config.Security.SecureEmbedEnabled {
            return HTTP 403 Forbidden
        }
    }
    ```

## Verification Plan
1.  **Unit Test:** Create a test in `internal/api/handlers/chatbot_unit_test.go` that attempts to enable secure embed for a Free plan user. Assert 403 Forbidden.
2.  **Manual Verification:** Use `curl` to send a PATCH request to a Free plan chatbot with `secure_embed_enabled: true`.
