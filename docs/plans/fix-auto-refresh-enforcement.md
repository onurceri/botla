# Plan: Fix Auto Refresh Enforcement

## Problem
The "Auto Refresh" feature (periodically recrawling sources) is restricted to paid plans. The Frontend hides the "Auto" option for Free plan users. However, the Backend API (`UpdateChatbot`) accepts `refresh_policy='auto'` without verifying if the user's plan allows it.

## Analysis
- **File:** `internal/api/handlers/chatbot_item.go`
- **Function:** `updateChatbot`
- **Current State:** The handler processes `RefreshPolicy` and sets `NextRefreshAt` if it is "auto", regardless of plan.
- **Requirement:** Validate against `PlanConfig.Refresh.Enabled`.

## Proposed Changes

### 1. Update `updateChatbot` Handler
- In `internal/api/handlers/chatbot_item.go`, before calling `applyChatbotUpdates`:
- Retrieve plan using `db.GetPlanByUserID`.
- Check if `req.RefreshPolicy` is "auto".
- If `plan.Config.Refresh.Enabled` is `false`, reject the request with 403 Forbidden.

### 2. Code Snippet
```go
if req.RefreshPolicy != nil && *req.RefreshPolicy == "auto" {
    plan, err := db.GetPlanByUserID(r.Context(), h.DB, c.UserID)
    if err != nil || plan == nil || !plan.Config.Refresh.Enabled {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Auto-refresh is not available on your plan",
            "upgrade_required": true,
        })
        return
    }
}
```

## Verification Plan
1.  **Unit Test:** Add test case in `internal/api/handlers/chatbot_unit_test.go` attempting to set `refresh_policy: "auto"` for a Free plan user.
2.  **Manual Verification:** Attempt to enable auto-refresh via API for a Free user.
